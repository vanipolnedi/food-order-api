package service

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"

	"food-order-api/internal/domain"
	"food-order-api/internal/promo"
	"food-order-api/internal/repository"
	"github.com/google/uuid"
)

var (
	ErrInvalidOrder    = errors.New("invalid order")
	ErrProductNotFound = errors.New("product not found")
)

type OrderService struct {
	productRepo repository.ProductRepository
	orderRepo   repository.OrderRepository
	promo       promo.Engine
}

// NewOrderService - constructs an OrderService with product, order, and promo dependencies.
func NewOrderService(productRepo repository.ProductRepository, orderRepo repository.OrderRepository, promoEngine promo.Engine) *OrderService {
	return &OrderService{
		productRepo: productRepo,
		orderRepo:   orderRepo,
		promo:       promoEngine,
	}
}

// Create - validates the request, prices items, applies a coupon if valid, and stores the order.
func (s *OrderService) Create(ctx context.Context, req domain.OrderRequest) (domain.Order, error) {
	if err := validateOrderRequest(req); err != nil {
		return domain.Order{}, err
	}
	/*
		Combine duplicate product IDs.
		Example:10 x 2, 10 x 3 becomes 10 x 5:
	*/
	quantities := make(map[string]int)
	for _, item := range req.Items {
		if strings.TrimSpace(item.ProductID) == "" {
			return domain.Order{}, fmt.Errorf(
				"%w: productId is required",
				ErrInvalidOrder,
			)
		}
		if item.Quantity <= 0 {
			return domain.Order{}, fmt.Errorf(
				"%w: quantity must be greater than zero",
				ErrInvalidOrder,
			)
		}
		// Protect against integer overflow.
		if quantities[item.ProductID] > math.MaxInt-item.Quantity {
			return domain.Order{}, fmt.Errorf(
				"%w: quantity overflow",
				ErrInvalidOrder,
			)
		}
		quantities[item.ProductID] += item.Quantity
	}
	items := make([]domain.OrderItem, 0, len(quantities))
	products := make([]domain.Product, 0, len(quantities))
	var subtotalCents int64
	for productID, quantity := range quantities {
		product, found, err := s.productRepo.Get(ctx, productID)
		if err != nil {
			return domain.Order{}, fmt.Errorf(
				"get product %s: %w",
				productID,
				err,
			)
		}
		if !found {
			return domain.Order{}, fmt.Errorf(
				"%w: %s",
				ErrProductNotFound,
				productID,
			)
		}
		priceCents := moneyToCents(product.Price)
		itemTotal := priceCents * int64(quantity)
		if itemTotal < 0 {
			return domain.Order{}, fmt.Errorf(
				"%w: price calculation overflow",
				ErrInvalidOrder,
			)
		}
		subtotalCents += itemTotal
		items = append(items, domain.OrderItem{
			ProductID: productID,
			Quantity:  quantity,
		})
		products = append(products, product)
	}
	// Invalid / unknown coupons do not fail the order. couponCode is
	// optional in the spec, so a bad code is treated as "no promo".
	discountCents := s.promo.Discount(ctx, req.CouponCode, subtotalCents)
	totalCents := subtotalCents - discountCents
	order := domain.Order{
		ID:        uuid.NewString(),
		Total:     centsToMoney(totalCents),
		Discounts: centsToMoney(discountCents),
		Items:     items,
		Products:  products,
	}
	if err := s.orderRepo.Create(ctx, order); err != nil {
		return domain.Order{}, fmt.Errorf(
			"create order: %w",
			err,
		)
	}
	return order, nil
}

// validateOrderRequest - rejects an order that has no items.
func validateOrderRequest(req domain.OrderRequest) error {
	if len(req.Items) == 0 {
		return fmt.Errorf(
			"%w: order must contain at least one item",
			ErrInvalidOrder,
		)
	}
	return nil
}

// moneyToCents - converts a dollar amount to integer cents.
func moneyToCents(value float64) int64 {
	return int64(math.Round(value * 100))
}

// centsToMoney - converts integer cents back to a dollar amount.
func centsToMoney(value int64) float64 {
	return float64(value) / 100
}
