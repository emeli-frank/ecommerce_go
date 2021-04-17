package postgres

import (
	"database/sql"
	"ecommerce/pkg/ecommerce"
	errors2 "ecommerce/pkg/ecommerce/errors"
	"ecommerce/pkg/storage"
	"fmt"
)

func NewOrderStorage(db *sql.DB) *orderStorage {
	return &orderStorage{db: db}
}

type orderStorage struct {
	db *sql.DB
}

func (s *orderStorage) SaveOrder(tx *sql.Tx, o *ecommerce.Order) (int, error) {
	const op = "orderStorage.SaveOrder"

	query := "INSERT INTO orders (product_id, ordered_at, shipping_address_id, quantity, customer_id) " +
		"VALUES ($1, $2, $3, $4, $5) RETURNING id"
	var id int
	err := tx.QueryRow(query, o.Product.ID, o.OrderedAt, o.ShippingAddressID, o.Quantity).Scan(&id)

	return id, errors2.Wrap(err, op, "executing query")
}

func (s *orderStorage) Order(id int) (*ecommerce.Order, error) {
	const op = "orderStorage.Order"

	query := fmt.Sprintf(
		`SELECT 
					product_id, 
					ordered_at, 
					shipping_address_id, 
					quantity, 
					customer_id 
				FROM orders JOIN PRODUC
				WHERE id = %d`,
		id,
	)

	var o ecommerce.Order
	o.ID = id
	err := s.db.QueryRow(query).Scan(&o.Product.ID, &o.OrderedAt, &o.ShippingAddressID, &o.Quantity, &o.CustomerID)
	if err == sql.ErrNoRows {
		return &o, errors2.Wrap(&errors2.NotFound{Err: err}, op, "executing query")
	}

	return &o, errors2.Wrap(err, op, "executing query")
}

func (s *orderStorage) Orders(ids []int) ([]ecommerce.Order, error) {
	const op = "orderStorage.Orders"

	if len(ids) < 1 {
		return nil, nil
	}

	query := fmt.Sprintf(
		`SELECT product_id, ordered_at, shipping_address_id, quantity, customer_id 
					WHERE id IN (%s)`,
		storage.IntSliceToCommaSeparatedStr(ids),
	) // todo:: order

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, errors2.Wrap(err, op, "executing query")
	}
	defer rows.Close()

	var oo []ecommerce.Order
	for rows.Next() {
		var o ecommerce.Order
		err := rows.Scan(&o.Product.ID, &o.OrderedAt, &o.ShippingAddressID, &o.Quantity, &o.CustomerID)
		if err == sql.ErrNoRows {
			return nil, errors2.Wrap(err, op, "scanning")
		}

		oo = append(oo, o)
	}

	return oo, errors2.Wrap(rows.Err(), op, "errors after row scan")
}

func (s *orderStorage) Tx() (*sql.Tx, error) {
	return s.db.Begin()
}
