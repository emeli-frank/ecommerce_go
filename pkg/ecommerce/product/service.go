package product

import (
	"database/sql"
	"ecommerce/pkg/ecommerce"
)

type repository interface {
	Products(
		categoryID int,
		searchTerm string,
		filter *ecommerce.ProductFilter,
		page int,
		size int) ([]ecommerce.Product, error)
	CreateCategory(name string) (int, error)
	CreateProduct(p *ecommerce.Product) (int, error)
}

func New(db *sql.DB, repo repository) *service {
	return &service{db: db, r: repo}
}

type service struct {
	db *sql.DB
	r repository
}

func (s *service) Products(
	categoryID int,
	searchTerm string,
	filter *ecommerce.ProductFilter,
	page int,
	size int) ([]ecommerce.Product, error) {
	return s.r.Products(categoryID, searchTerm, filter, page, size)
}

func (s *service) CreateCategory(name string) (int, error) {
	return s.r.CreateCategory(name)
}
func (s *service) CreateProduct(p *ecommerce.Product) (int, error) {
	return s.r.CreateProduct(p)
}
