package user

import (
	"database/sql"
	"ecommerce/pkg/ecommerce"
	errors2 "ecommerce/pkg/ecommerce/errors"
	"golang.org/x/crypto/bcrypt"
)

type repository interface {
	SaveUserWithTx(tx *sql.Tx, user *ecommerce.User, hashedPassword string) (int, error)
	UpdateRolesWithTx(tx *sql.Tx, uid int, roles []int) error
	UserIDAndPasswordByEmail(email string) (int, string, error)
	User(uid int) (*ecommerce.User, error)
	UpdateUser(user *ecommerce.User) error
	Tx() (*sql.Tx, error)
}

func New(db *sql.DB, repo repository) *service {
	return &service{db: db, r: repo}
}

type service struct {
	db *sql.DB
	r repository
}

func (s *service) CreateCustomer(c *ecommerce.User, password string) (int, error) {
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
	id, err := s.r.SaveUserWithTx(tx, c, string(hash))
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

	return id, errors2.Wrap(tx.Commit(), op, "committing tx")
}

func (s *service) EmailMatchPassword(email string, password string) (bool, int, error) {
	op := "userService.EmailMatchPassword"

	// todo:: validate email and password

	uid, hashedPassword, err := s.r.UserIDAndPasswordByEmail(email)
	if err != nil {
		switch errors2.Unwrap(err).(type) {
		case *errors2.NotFound:
			return false, 0, nil
		default:
			return false, 0, err
		}
	}

	// compare user provided and stored password
	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err == bcrypt.ErrMismatchedHashAndPassword {
		return false, 0, nil
	} else if err != nil {
		return false, 0, errors2.Wrap(err, op, "hashing password")
	}

	return true, uid, nil
}

func (s *service) User(uid int) (*ecommerce.User, error) {
	const op = "userService.User"

	u, err := s.r.User(uid)
	if err != nil {
		return nil, errors2.Wrap(err, op, "getting user from repo")
	}

	return u, nil
}

func (s *service) UpdateUser(user *ecommerce.User) error {
	const op = "userService.UpdateUser"

	return errors2.Wrap(s.r.UpdateUser(user), op, "updating from repo")
}
