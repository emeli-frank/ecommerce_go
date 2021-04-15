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

func (s *userStorage) SaveUserWithTx(tx *sql.Tx, user *ecommerce.User, hashedPassword string) (int, error) {
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

func (s *userStorage) UpdateUser(user *ecommerce.User) error {
	const op = "userStorage.UpdateUser"

	query := "UPDATE users SET " +
		"first_name = $1," +
		"last_name = $2," +
		"email = $3" +
		"WHERE id = $4"
	_, err := s.db.Exec(query, user.FirstName, user.LastName, user.Email, user.ID)
	if err != nil {
		return errors2.Wrap(err, op, "executing query")
	}

	return nil
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

func (s *userStorage) UserIDAndPasswordByEmail(email string) (int, string, error) {
	const op = "userStorage.UserIDAndPasswordByEmail"

	query := `SELECT id, password FROM users WHERE email = $1`

	var id int
	var password string
	err := s.db.QueryRow(query, email).Scan(&id, &password)
	if err == sql.ErrNoRows {
		err = &errors2.NotFound{Err: errors.New("user not found")}
		return 0, "", errors2.Wrap(err, op, "scanning into var")
	} else if err != nil {
		return 0, "", errors2.Wrap(err, op, "scanning into var")
	}

	return id, password, nil
}

func (s userStorage) User(uid int) (*ecommerce.User, error) {
	const op = "userStorage.User"

	query := fmt.Sprintf(`SELECT 
				users.id, 
				users.first_name, 
				users.last_name, 
				users.email,
				role_user_map.role_id
			FROM users
			INNER JOIN role_user_map ON users.id = role_user_map.user_id
			WHERE users.id = %d`, uid)

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, errors2.Wrap(err, op, "querying rows")
	}
	defer rows.Close()

	var u *ecommerce.User
	var r int

	if rows.Next() {
		tempUser := ecommerce.User{}
		err = rows.Scan(&tempUser.ID, &tempUser.FirstName, &tempUser.LastName, &tempUser.Email, &r)
		if err != nil {
			return nil, errors2.Wrap(err, op, "scanning into struct")
		}
		tempUser.Roles = append(tempUser.Roles, r)
		u = &tempUser
	} else {
		return nil, errors2.Wrap(&errors2.NotFound{Err:errors.New("user not found")}, op, "")
	}

	for rows.Next() {
		var dummyVar interface{}
		err = rows.Scan(&dummyVar, &dummyVar, &dummyVar, &dummyVar, &r)
		if err != nil {
			return nil, errors2.Wrap(err, op, "scanning into dummy var and role var")
		}
		u.Roles = append(u.Roles, r)
	}

	if err = rows.Err(); err != nil {
		return nil, errors2.Wrap(err, op, "checking error after iterating rows.Next()")
	}

	return u, nil
}
