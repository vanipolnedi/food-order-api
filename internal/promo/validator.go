package promo

import (
	"bufio"
	"compress/gzip"
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"unicode"
)

const (
	minCodeLength = 8
	maxCodeLength = 10
	minFileHits   = 2
	// Production coupon dumps can have long noisy lines. Keep a generous
	// scan buffer so we do not fail on the first oversized token.
	scannerMaxToken = 1 << 20
)

// Checker is the only coupon API the order flow needs.
// A later Redis / DB implementation can replace FileIndex without
// touching handlers.
type Checker interface {
	Validate(ctx context.Context, code string) bool
}

// FileIndex loads gzip coupon dumps once, then answers Validate in O(1).
type FileIndex struct {
	mu     sync.RWMutex
	counts map[string]int
}

// NewValidator - constructs an empty in-memory coupon file index.
func NewValidator() *FileIndex {
	return &FileIndex{
		counts: make(map[string]int),
	}
}

// LoadFiles - streams gzip coupon dumps in parallel and builds code-to-file-count index.
func (v *FileIndex) LoadFiles(ctx context.Context, paths []string) error {
	type fileResult struct {
		codes map[string]struct{}
		err   error
	}
	results := make([]fileResult, len(paths))
	var wg sync.WaitGroup
	for i, path := range paths {
		wg.Add(1)
		go func(i int, path string) {
			defer wg.Done()
			codes, err := readFile(ctx, path)
			results[i] = fileResult{codes: codes, err: err}
		}(i, path)
	}
	wg.Wait()
	counts := make(map[string]int)
	for _, result := range results {
		if result.err != nil {
			return result.err
		}
		for code := range result.codes {
			counts[code]++
		}
	}
	v.mu.Lock()
	v.counts = counts
	v.mu.Unlock()
	return nil
}

// Validate - returns true when the code is 8-10 chars and appears in at least two files.
func (v *FileIndex) Validate(ctx context.Context, code string) bool {
	if err := ctx.Err(); err != nil {
		return false
	}
	code = strings.TrimSpace(code)
	if !validLength(code) {
		return false
	}
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.counts[code] >= minFileHits
}

// readFile - decompresses one gzip coupon file and returns unique codes found in it.
func readFile(ctx context.Context, path string) (map[string]struct{}, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open coupon file %s: %w", path, err)
	}
	defer file.Close()
	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		return nil, fmt.Errorf("create gzip reader for %s: %w", path, err)
	}
	defer gzipReader.Close()
	codes := make(map[string]struct{})
	scanner := bufio.NewScanner(gzipReader)
	scanner.Buffer(make([]byte, 0, 64*1024), scannerMaxToken)
	for scanner.Scan() {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		collectCodes(scanner.Text(), codes)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read coupon file %s: %w", path, err)
	}
	return codes, nil
}

// collectCodes - extracts unique 8-10 character alphanumeric tokens from a line.
func collectCodes(line string, codes map[string]struct{}) {
	start := -1
	flush := func(end int) {
		if start < 0 {
			return
		}
		token := line[start:end]
		if validLength(token) {
			codes[token] = struct{}{}
		}
		start = -1
	}
	for i, r := range line {
		if isCodeRune(r) {
			if start < 0 {
				start = i
			}
			continue
		}
		flush(i)
	}
	flush(len(line))
}

// isCodeRune - returns true when the rune is a letter or digit.
func isCodeRune(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r)
}

// validLength - returns true when the code length is between 8 and 10 characters.
func validLength(code string) bool {
	n := len(code)
	return n >= minCodeLength && n <= maxCodeLength
}
