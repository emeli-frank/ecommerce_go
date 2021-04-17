package postgres

import (
	"database/sql"
	"ecommerce/pkg/ecommerce"
	errors2 "ecommerce/pkg/ecommerce/errors"
	"errors"
	"fmt"
)

func NewAddressStorage(db *sql.DB) *addressStorage {
	return &addressStorage{db: db}
}

type addressStorage struct {
	db *sql.DB
}

func (s *addressStorage) SaveAddressWithTx(tx *sql.Tx, a *ecommerce.Address) (int, error) {
	const op = "userStorage.SaveAddressWithTx"

	if tx == nil {
		return 0, errors2.Wrap(errors.New("transaction is nil"), op, "")
	}

	query := "INSERT INTO addresses (country, state, city, postal_code, address) VALUES ($1, $2, $3, $4, $5) RETURNING id"
	var id int
	err := tx.QueryRow(query, a.Country, a.State, a.City, a.PostalCode, a.Address).Scan(&id)

	return id, errors2.Wrap(err, op, "executing query")
}

func (s *addressStorage) UpdateAddress(a *ecommerce.Address) error {
	const op = "userStorage.UpdateAddress"

	query := "UPDATE addresses SET country = $1, state = $2, city = $3, postal_code = $4, address = $5 WHERE id = $6"
	_, err := s.db.Exec(query, a.Country, a.State, a.City, a.PostalCode, a.Address, a.ID)

	return errors2.Wrap(err, op, "executing query")
}

func (s *addressStorage) Address(id int) (*ecommerce.Address, error) {
	const op = "userStorage.Address"

	query := fmt.Sprintf("SELECT country, state, city, postal_code, address FROM addresses WHERE id = %d", id)

	var a ecommerce.Address
	a.ID = id
	err := s.db.QueryRow(query).Scan(&a.Country, &a.State, &a.City, &a.PostalCode, &a.Address)
	if err == sql.ErrNoRows {
		return &a, errors2.Wrap(&errors2.NotFound{Err: err}, op, "executing query")
	}

	return &a, errors2.Wrap(err, op, "executing query")
}

func (s *addressStorage) DeleteAddress(tx *sql.Tx, id int) error {
	const op = "userStorage.DeleteAddress"

	query := fmt.Sprintf("DELETE FROM addresses WHERE id = %d", id)
	fmt.Println(query)
	_, err := s.db.Exec(query)

	return errors2.Wrap(err, op, "executing query")
}

func (s *addressStorage) Tx() (*sql.Tx, error) {
	return s.db.Begin()
}
