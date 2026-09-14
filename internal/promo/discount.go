package promo

import (
	"context"
	"math"
	"strings"
)

// DiscountFunc turns a coupon + subtotal into a discount in cents.
// Swap this later for per-code rules (FIFTYOFF = 50%, HAPPYHRS = happy-hour %).
type DiscountFunc func(subtotalCents int64, code string) int64

// PercentOff - returns a discount function that takes the given percent off the subtotal.
func PercentOff(percent float64) DiscountFunc {
	return func(subtotalCents int64, _ string) int64 {
		if percent <= 0 || subtotalCents <= 0 {
			return 0
		}
		return int64(math.Round(float64(subtotalCents) * percent))
	}
}

// Engine is what OrderService depends on. Today it is a file-backed
// checker plus a flat percent. Tomorrow it can be a campaign service.
type Engine interface {
	Discount(ctx context.Context, code string, subtotalCents int64) int64
}

type FileEngine struct {
	checker  Checker
	discount DiscountFunc
}

// NewFileEngine - constructs a promo engine from a coupon checker and discount function.
func NewFileEngine(checker Checker, discount DiscountFunc) *FileEngine {
	return &FileEngine{
		checker:  checker,
		discount: discount,
	}
}

// Discount - returns the discount in cents for a coupon, or 0 when the code is invalid.
func (e *FileEngine) Discount(ctx context.Context, code string, subtotalCents int64) int64 {
	code = strings.TrimSpace(code)
	if code == "" {
		return 0
	}
	if !e.checker.Validate(ctx, code) {
		return 0
	}
	amount := e.discount(subtotalCents, code)
	if amount < 0 {
		return 0
	}
	if amount > subtotalCents {
		return subtotalCents
	}
	return amount
}
