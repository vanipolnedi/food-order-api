package repository

import (
	"context"
	"sync"

	"food-order-api/internal/domain"
)

type OrderRepository interface {
	Create(ctx context.Context, order domain.Order) error
}

type InMemoryOrderRepository struct {
	mu     sync.RWMutex
	orders map[string]domain.Order
}

// NewInMemoryOrderRepository - constructs an empty in-memory order store.
func NewInMemoryOrderRepository() *InMemoryOrderRepository {
	return &InMemoryOrderRepository{
		orders: make(map[string]domain.Order),
	}
}

// Create - stores an order in memory by id.
func (r *InMemoryOrderRepository) Create(ctx context.Context, order domain.Order) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.orders[order.ID] = order
	return nil
}
