# Food Ordering API

Go server for a food-ordering challenge. Spec: `openapi.yaml`.

**Problem.** Build a robust Go API that matches the OpenAPI 3.1 spec as closely
as possible. A client must be able to list the menu, fetch one product, and
place an order (optional promo). `POST /order` requires header `api_key: apitest`.
Promo codes live in three gzip dumps (`data/couponbase1.gz`, `couponbase2.gz`, `couponbase3.gz`). A code is valid only if it is 8–10 characters and appears in **at least two** of those files. Examples: `HAPPYHRS` and `FIFTYOFF` are valid `SUPER100` is not. Handle edge cases the demo API skips.

The interesting part of this implementation is not the three HTTP routes. It is
how coupons are indexed once at startup, then reused on every order without
rereading the dumps.

## What the API does

| Method | Path | Auth | Result |
| ---  | --- | --- | --- |
| GET  | `/product` | none | full menu |
| GET  | `/product/{productId}` | none | one item, or 400 / 404 |
| POST | `/order` | header `api_key: apitest` | create order, optional coupon |

Missing key → 401. Wrong key → 403. Bad JSON → 400. Empty cart / unknown
product → 422.

## Coupon rules

A code is valid only when all of these are true:

1. Length is 8, 9 or 10 characters.
2. It appears in **at least two different** gzip files.
3. Repeats inside the **same** file count as one hit.

Checked-in samples (small stand-ins for the original large dumps):

| Code | Files | Result |
| --- | --- | --- |
| `HAPPYHRS` | 1 and 2 | valid, 10% off |
| `FIFTYOFF` | 2 and 3 | valid, 10% off |
| `SUPER100` | 1 only | invalid, order still placed, `$0` discount |

`couponCode` is optional in the spec, so an invalid code does not fail the
order. It is treated as “no promo”. If product later wants a hard 422, that
change lives in `promo.FileEngine`, not in the HTTP handler.

The spec never defines the discount amount. This build uses a flat 10% so
orders have a deterministic `discounts` / `total`. Per-code amounts
(`FIFTYOFF` → 50%) can be plugged in through `promo.DiscountFunc`.

The real coupon dumps are hundreds of MB each (not committed). Download them
into `data/` before running against production-sized files:

```bash
curl -fL -o data/couponbase1.gz \
  https://orderfoodonline-files.s3.ap-southeast-2.amazonaws.com/couponbase1.gz
curl -fL -o data/couponbase2.gz \
  https://orderfoodonline-files.s3.ap-southeast-2.amazonaws.com/couponbase2.gz
curl -fL -o data/couponbase3.gz \
  https://orderfoodonline-files.s3.ap-southeast-2.amazonaws.com/couponbase3.gz
```

## How validation scales

The original dumps are described as large gzip text files. Scanning them on
every `POST /order` would not survive traffic.

Startup:

1. Open the three gzip files **in parallel**.
2. Stream each file line by line. Do not hold the decompressed dump.
3. Keep only alphanumeric tokens of length 8–10. That covers both
   “one code per line” samples and noisy “random text” dumps.
4. Collapse duplicates per file, then count how many files contain each code.
5. Store `map[code]fileCount` in memory.

Request path:

`Validate` is a map lookup under a read lock. No disk I/O.

Memory is proportional to **unique candidate codes**, not file size. Three
large files with lots of repeated junk stay small. If the set no longer fits
on one box, the `promo.Checker` interface is the swap point:

- Redis `SET` / HyperLogLog per file, or
- object store + periodic rebuild of the index, or
- a campaign service that already knows valid codes.

Do not index every sliding 8–10 gram of a multi-GB random blob. That grows
with file size and will not fit in RAM. Tokenizing on non-alphanumeric
boundaries is the practical tradeoff for this problem.

## Layout

```
internal/handler      HTTP only. Status codes, JSON, no coupon math.
internal/service      Order / product rules. Talks to interfaces.
internal/promo        Coupon index + discount engine.
internal/repository   In-memory stores. Same interfaces can wrap SQL later.
internal/middleware   API key, request id, logs, panic recovery.
internal/domain       Request/response models.
cmd/server            Process entrypoint.
data                  Sample gzip coupon files.
```

In-memory products and orders are enough for the challenge. For a real
catalog / order history, replace the repository implementations. Handlers
do not change.

## Run with Go

Needs Go 1.22+. From the repository root (the folder that contains `go.mod`):

```bash
go run ./cmd/server
```

The server listens on `:8080`. Coupon files default to `./data`. Override with
`COUPON_DATA_DIR` if you run the binary from another directory.

Imports use the module name in `go.mod` (`food-order-api/...`), not Uber paths.
Anyone who clones this repo can `go run` without extra GOPATH setup.

If you publish to GitHub and want `go install github.com/<you>/food-order-api/...`
to work, change the first line of `go.mod` to that path and rename the
`food-order-api/...` imports to match. Local `go run ./cmd/server` does not
need that.

## Run with Docker

All commands below run from the repository root (the folder with `Dockerfile`):

### Option A: plain Docker (works everywhere)

```bash
docker build -t food-order-api .
docker run --rm -p 8080:8080 --name food-order-api food-order-api
```

Stop with Ctrl+C, or from another shell:

```bash
docker stop food-order-api
```

### Option B: Compose

Compose ships in two forms. Use whichever your machine has:

```bash
# Compose v2 (plugin, newer Docker Desktop)
docker compose up --build

# Compose v1 (standalone binary, note the hyphen)
docker-compose up --build
```

If `docker compose up --build` fails with `unknown flag: --build`, your Docker
CLI has no Compose plugin. Use the hyphenated `docker-compose` command, or
Option A.

Check which one you have:

```bash
docker compose version || docker-compose version
```

Stop with Ctrl+C, then:

```bash
docker compose down or docker-compose down
```

### Verify it works

```bash
curl -s localhost:8080/product

curl -s localhost:8080/product/10

# valid coupon: 10% off -> discounts 2.66, total 23.94
curl -s -X POST localhost:8080/order \
  -H 'Content-Type: application/json' \
  -H 'api_key: apitest' \
  -d '{"couponCode":"HAPPYHRS","items":[{"productId":"10","quantity":2}]}'

# invalid coupon: order still placed, discounts 0, total 26.60
curl -s -X POST localhost:8080/order \
  -H 'Content-Type: application/json' \
  -H 'api_key: apitest' \
  -d '{"couponCode":"SUPER100","items":[{"productId":"10","quantity":2}]}'

# missing api_key -> 401
curl -s -o /dev/null -w '%{http_code}\n' -X POST localhost:8080/order \
  -d '{"items":[{"productId":"10","quantity":1}]}'
```

The image already sets `COUPON_DATA_DIR=/app/data` and copies the coupon
files in, so no extra setup is needed.

## Assumptions worth calling out

- Menu is seeded in `main.go`. The spec examples a waffle, that is the catalog.
- Valid coupon → 10% off the subtotal, rounded to cents.
- Invalid coupon → 200 + no discount, not 422.
- Product path param is an int64 at the HTTP boundary. product `id` is a string
  in JSON, matching the spec's mixed types.
