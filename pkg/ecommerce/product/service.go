package product

import (
	"database/sql"
	"ecommerce/pkg/ecommerce"
	"ecommerce/pkg/ecommerce/errors"
)

type repository interface {
	ProductIDs(
		categoryID int,
		searchTerm string,
		filter *ecommerce.ProductFilter,
		page int,
		size int) ([]int, error)
	ProductsFromIDs(ids []int) ([]ecommerce.Product, error)
	Product(id int) (*ecommerce.Product, error)
	CreateCategory(name string) (int, error)
	CreateProduct(p *ecommerce.Product) (int, error)
	UpdateProductWithTx(tx *sql.Tx, p *ecommerce.Product) error
	Tx() (*sql.Tx, error)
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
	const op = "productService.Products"

	ids, err := s.r.ProductIDs(categoryID, searchTerm, filter, page, size)
	if err != nil {
		return nil, errors.Wrap(err, op, "getting product ids")
	}

	pp, err := s.ProductsFromIDs(ids)

	return pp, errors.Wrap(err, op, "getting products from ids")
}

func (s *service) CreateCategory(name string) (int, error) {
	return s.r.CreateCategory(name)
}
func (s *service) CreateProduct(p *ecommerce.Product) (int, error) {
	return s.r.CreateProduct(p)
}

func (s *service) UpdateProductWithTx(tx *sql.Tx, p *ecommerce.Product) error {
	const op = "productService.UpdateProductWithTx"

	tx, err := s.r.Tx()
	if err != nil {
		return errors.Wrap(err, op, "getting tx")
	}

	err = s.r.UpdateProductWithTx(tx, p)
	if err != nil {
		return errors.Wrap(err, op, "updating product")
	}

	return errors.Wrap(tx.Commit(), op, "committing tx")
}

func (s *service) Product(id int) (*ecommerce.Product, error) {
	const op  = "service.Product"

	p, err := s.r.Product(id)

	return p, errors.Wrap(err, op, "getting product from repo")
}

func (s *service) ProductsFromIDs(ids []int) ([]ecommerce.Product, error) {
	const op = "userService.ProductsFromID"

	pp, err := s.r.ProductsFromIDs(ids)

	return pp, errors.Wrap(err, op, "getting product ids from repo")
}
