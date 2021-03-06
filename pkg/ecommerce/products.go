package ecommerce

import "database/sql"

type ProductService interface {
	Products(
		categoryID int,
		searchTerm string,
		filter *ProductFilter,
		page int,
		size int) ([]Product, error)
	CreateCategory(name string) (int, error)
	CreateProduct(p *Product) (int, error)
	UpdateProductWithTx(tx *sql.Tx, p *Product) error
	Product(id int) (*Product, error)
	ProductsFromIDs(ids []int) ([]Product, error)
}

type Product struct {
	ID int `json:"id"`
	Name string `json:"name"`
	CategoryID int `json:"category_id"`
	Price Price `json:"price"`
	Rating int `json:"rating,omitempty"`
	Description string `json:"description,omitempty"`
	Quantity int `json:"quantity,omitempty"`
}

type Category struct {
	ID int `json:"id"`
	Name string `json:"name"`
}
