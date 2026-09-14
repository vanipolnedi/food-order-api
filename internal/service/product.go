package service

import (
	"context"

	"food-order-api/internal/domain"
	"food-order-api/internal/repository"
)

type ProductService struct {
	repo repository.ProductRepository
}

// NewProductService - constructs a ProductService with the given repository.
func NewProductService(repo repository.ProductRepository) *ProductService {
	return &ProductService{
		repo: repo,
	}
}

// List - returns every product in the catalog.
func (s *ProductService) List(ctx context.Context) ([]domain.Product, error) {
	return s.repo.List(ctx)
}

// Get - returns a product by id and whether it exists.
func (s *ProductService) Get(ctx context.Context, id string) (domain.Product, bool, error) {
	return s.repo.Get(ctx, id)
}
