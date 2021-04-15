package postgres

import (
	"database/sql"
	"ecommerce/pkg/ecommerce"
	errors2 "ecommerce/pkg/ecommerce/errors"
	"errors"
	"fmt"
)

func NewUserStorage(db *sql.DB) *userStorage {
	return &userStorage{db: db}
}

type userStorage struct {
	db *sql.DB
}

func (s *userStorage) SaveUserWithTx(tx *sql.Tx, user *ecommerce.UserBase, hashedPassword string) (int, error) {
	const op = "userStorage.SaveUserWithTx"

	if tx == nil {
		return 0, errors2.Wrap(errors.New("transaction is nil"), op, "")
	}

	query := "INSERT INTO users (first_name, last_name, email, password) VALUES ($1, $2, $3, $4) RETURNING id"
	var id int
	err := tx.QueryRow(query, user.FirstName, user.LastName, user.Email, hashedPassword).Scan(&id)
	if err != nil {
		return 0, errors2.Wrap(err, op, "executing query")
	}

	return id, nil
}

func (s *userStorage) UpdateRolesWithTx(tx *sql.Tx, uid int, roles []int) error {
	const op = "userStorage.UpdateRolesWithTx"

	if tx == nil {
		return errors2.Wrap(errors.New("transaction is nil"), op, "")
	}

	// delete all user roles
	err := s.deleteAllRoles(uid, tx)
	if err != nil {
		return errors2.Wrap(err, op, "deleting roles")
	}

	// attach new roles
	err = s.attachRoles(uid, roles, tx)
	if err != nil {
		return errors2.Wrap(err, op, "attaching new roles roles")
	}

	return nil
}

func (s *userStorage) attachRoles(userId int, roles []int, tx *sql.Tx) error {
	const op = "userStorage.attachRoles"

	rolesCount := len(roles)
	if rolesCount < 1 {
		return nil
	}

	query := "INSERT INTO role_user_map (user_id, role_id) VALUES"

	for k, roleId := range roles {
		query += fmt.Sprintf("(%d, %d)", userId, roleId)
		if k < rolesCount - 1 {
			query += ","
		}
	}

	_, err := tx.Exec(query)
	if err != nil {
		return errors2.Wrap(err, op, "executing query")
	}

	return nil
}

func (s *userStorage) deleteAllRoles(uid int, tx *sql.Tx) error {
	const op = "userStorage.deleteAllRoles"

	query := fmt.Sprintf("DELETE FROM role_user_map WHERE user_id = %d", uid)

	_, err := tx.Exec(query)
	if err != nil {
		return errors2.Wrap(err, op, "executing query")
	}

	return nil
}

func (s *userStorage) Tx() (*sql.Tx, error) {
	return s.db.Begin()
}
