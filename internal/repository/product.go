package repository

import (
	"context"
	"sync"

	"food-order-api/internal/domain"
)

type ProductRepository interface {
	List(ctx context.Context) ([]domain.Product, error)
	Get(ctx context.Context, id string) (domain.Product, bool, error)
}

type InMemoryProductRepository struct {
	mu       sync.RWMutex
	products map[string]domain.Product
}

// NewInMemoryProductRepository - constructs an in-memory product store from the given catalog.
func NewInMemoryProductRepository(products []domain.Product) *InMemoryProductRepository {
	productMap := make(map[string]domain.Product, len(products))
	for _, product := range products {
		productMap[product.ID] = product
	}
	return &InMemoryProductRepository{
		products: productMap,
	}
}

// List - returns a copy of every product in the in-memory catalog.
func (r *InMemoryProductRepository) List(ctx context.Context) ([]domain.Product, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]domain.Product, 0, len(r.products))
	for _, product := range r.products {
		result = append(result, product)
	}
	return result, nil
}

// Get - returns a product by id from the in-memory catalog.
func (r *InMemoryProductRepository) Get(ctx context.Context, id string) (domain.Product, bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	product, ok := r.products[id]
	return product, ok, nil
}
