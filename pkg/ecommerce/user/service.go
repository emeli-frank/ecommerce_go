package user

import (
	"database/sql"
	"ecommerce/pkg/ecommerce"
	"ecommerce/pkg/ecommerce/errors"
	"golang.org/x/crypto/bcrypt"
)

type repository interface {
	SaveUserWithTx(tx *sql.Tx, user *ecommerce.UserBase, hashedPassword string) (int, error)
	UpdateRolesWithTx(tx *sql.Tx, uid int, roles []int) error
	Tx() (*sql.Tx, error)
}

func New(db *sql.DB, repo repository) *service {
	return &service{db: db, r: repo}
}

type service struct {
	db *sql.DB
	r repository
}

func (s *service) CreateCustomer(c *ecommerce.Customer, password string) (int, error) {
	const op = "userService.CreateCustomer"

	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return 0, err
	}

	tx, err := s.r.Tx()
	if err != nil {
		return 0, err
	}

	// create user
	id, err := s.r.SaveUserWithTx(tx, &c.UserBase, string(hash))
	if err != nil {
		_ = tx.Rollback()
		return 0, err
	}

	// update role
	err = s.r.UpdateRolesWithTx(tx, id, []int{ecommerce.RoleCustomer})
	if err != nil {
		_ = tx.Rollback()
		return 0, err
	}

	return id, errors.Wrap(tx.Commit(), op, "committing tx")
}
