package postgres

import (
	"database/sql"
	"ecommerce/pkg/ecommerce"
	errors2 "ecommerce/pkg/ecommerce/errors"
	"ecommerce/pkg/storage"
	"errors"
	"fmt"
)

func NewProductStorage(db *sql.DB) *productStorage {
	return &productStorage{db: db}
}

type productStorage struct {
	db *sql.DB
}

/*func (s *productStorage) Products(
	categoryID int,
	searchTerm string,
	filter *ecommerce.ProductFilter,
	page int,
	size int) ([]ecommerce.Product, error) {

	var params []interface{}
	var categoryQuery string
	if categoryID > 0 {
		categoryQuery = fmt.Sprintf("AND category_id = %d", categoryID)
	}

	var searchQuery string
	if searchTerm != "" {
		searchQuery = "AND name LIKE $1"
		params = append(params, "%" + searchTerm + "%")
	}

	var minPriceQuery, maxPriceQuery, discountQuery string
	if filter != nil {
		if filter.MinPrice > 0 {
			minPriceQuery = fmt.Sprintf("AND price >= %f", filter.MinPrice)
		}
		if filter.MaxPrice > 0 {
			maxPriceQuery = fmt.Sprintf("AND price <= %f", filter.MaxPrice)
		}
		//if filter.Discount > 0 {
		//	discountQuery = "AND p.price"
		//}
	}

	var limitQuery string
	if page < 1 {
		return nil, errors.New("page cannot be less than one")
	}

	if size < 1 {
		return nil, errors.New("size cannot be less than one")
	}

	offset := (page - 1) * size
	limitQuery = fmt.Sprintf("LIMIT %d OFFSET %d", size, offset)

	query := fmt.Sprintf("SELECT id, category_id, name, price, old_price, rating FROM products " +
		"WHERE 1=1 %s %s %s %s %s ORDER BY id %s",
		categoryQuery, searchQuery, minPriceQuery, maxPriceQuery, discountQuery, limitQuery)

	fmt.Println(query)
	fmt.Println("page", page)

	row, err := s.db.Query(query, params...)
	if err != nil {
		return nil, err
	}
	defer row.Close()

	var pp []ecommerce.Product
	for row.Next() {
		var p ecommerce.Product
		var oldPrice sql.NullFloat64
		var rating sql.NullInt64
		err = row.Scan(&p.ID, &p.CategoryID, &p.Name, &p.Price.Current, &oldPrice, &rating)
		if err != nil {
			return nil, err
		}
		p.Price.Old = float32(storage.NullableFloatToFloat(oldPrice))
		p.Rating = int(storage.NullableIntToInt(rating))
		pp = append(pp, p)
	}

	if err = row.Err(); err != nil {
		return nil, err
	}

	return pp, nil
}*/

func (s *productStorage) ProductIDs(
	categoryID int,
	searchTerm string,
	filter *ecommerce.ProductFilter,
	page int,
	size int) ([]int, error) {

	const op = "productStorage.ProductIDs"

	var params []interface{}
	var categoryQuery string
	if categoryID > 0 {
		categoryQuery = fmt.Sprintf("AND category_id = %d", categoryID)
	}

	var searchQuery string
	if searchTerm != "" {
		searchQuery = "AND name LIKE $1"
		params = append(params, "%" + searchTerm + "%")
	}

	var minPriceQuery, maxPriceQuery, discountQuery string
	if filter != nil {
		if filter.MinPrice > 0 {
			minPriceQuery = fmt.Sprintf("AND price >= %f", filter.MinPrice)
		}
		if filter.MaxPrice > 0 {
			maxPriceQuery = fmt.Sprintf("AND price <= %f", filter.MaxPrice)
		}
		/*if filter.Discount > 0 {
			discountQuery = "AND p.price"
		}*/
	}

	var limitQuery string
	if page < 1 {
		return nil, errors2.Wrap(errors.New("page cannot be less than one"), op, "")
	}

	if size < 1 {
		return nil, errors.New("size cannot be less than one")
	}

	offset := (page - 1) * size
	limitQuery = fmt.Sprintf("LIMIT %d OFFSET %d", size, offset)

	query := fmt.Sprintf("SELECT id FROM products WHERE 1=1 %s %s %s %s %s ORDER BY id %s",
		categoryQuery, searchQuery, minPriceQuery, maxPriceQuery, discountQuery, limitQuery)

	fmt.Println(query)
	fmt.Println("page", page)

	row, err := s.db.Query(query, params...)
	if err != nil {
		return nil, errors2.Wrap(err, op, "executing query")
	}
	defer row.Close()

	var ids []int
	for row.Next() {
		var id int
		err = row.Scan(&id)
		if err != nil {
			return nil, errors2.Wrap(err, op, "scanning")
		}
		ids = append(ids, id)
	}

	return ids, errors2.Wrap(row.Err(), op, "error after scan")
}

func (s *productStorage) ProductsFromIDs(ids []int) ([]ecommerce.Product, error) {

	if len(ids) < 1 {
		return nil, nil
	}

	query := fmt.Sprintf("SELECT id, category_id, name, price, old_price, rating FROM products WHERE id IN (%s)", storage.IntSliceToCommaSeparatedStr(ids))

	row, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer row.Close()

	var pp []ecommerce.Product
	for row.Next() {
		var p ecommerce.Product
		var oldPrice sql.NullFloat64
		var rating sql.NullInt64
		err = row.Scan(&p.ID, &p.CategoryID, &p.Name, &p.Price.Current, &oldPrice, &rating)
		if err != nil {
			return nil, err
		}
		p.Price.Old = float32(storage.NullableFloatToFloat(oldPrice))
		p.Rating = int(storage.NullableIntToInt(rating))
		pp = append(pp, p)
	}

	if err = row.Err(); err != nil {
		return nil, err
	}

	return pp, nil
}

func (s *productStorage) CreateCategory(name string) (int, error) {
	query := "INSERT INTO product_categories (name) VALUES ($1) RETURNING id"

	var id int
	err := s.db.QueryRow(query, name).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (s *productStorage) CreateProduct(p *ecommerce.Product) (int, error) {
	query := "INSERT INTO products (name, category_id, price, description, quantity) " +
		"VALUES ($1, $2, $3, $4, $5) RETURNING id"

	var id int
	err := s.db.QueryRow(query, p.Name, p.CategoryID, p.Price.Current, p.Description, p.Quantity).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (s *productStorage) Product(id int) (*ecommerce.Product, error) {
	const op  = "productStorage.Product"

	query := fmt.Sprintf("SELECT name, category_id, price, description, quantity FROM products WHERE id = %d", id)

	var p ecommerce.Product
	p.ID = id
	err := s.db.QueryRow(query).Scan(&p.Name, &p.CategoryID, &p.Price.Current, &p.Description, &p.Quantity)

	return &p, errors2.Wrap(err, op, "executing query")
}

func (s *productStorage) UpdateProductWithTx(tx *sql.Tx, p *ecommerce.Product) error {
	const op = "productStorage.UpdateProductWithTx"

	query := "UPDATE products SET name = $1, category_id = $2, price = $3, description = $4, quantity = $5 WHERE id = $6"
	_, err := tx.Exec(query, p.Name, p.CategoryID, p.Price.Current, p.Description, p.Quantity)

	return errors2.Wrap(err, op, "executing query")
}

func (s *productStorage) Tx() (*sql.Tx, error) {
	return s.db.Begin()
}
