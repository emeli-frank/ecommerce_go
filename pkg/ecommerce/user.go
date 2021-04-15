package ecommerce

import (
	"errors"
	"github.com/dgrijalva/jwt-go"
	"time"
)

const (
	RoleCustomer = 1
)
type UserService interface {
	CreateCustomer(c *User, password string) (int, error)
	EmailMatchPassword(email string, password string) (bool, int, error)
	User(uid int) (*User, error)
	UpdateUser(user *User) error
}

type UserClaims struct {
	UserID         int   `json:"user_id"`
	Roles          []int `json:"roles"`
	jwt.StandardClaims
}

type User struct {
	ID int `json:"id"`
	FirstName string `json:"first_name"`
	LastName string `json:"last_name"`
	Email string `json:"email"`
	Roles []int `json:"roles"`
}

func (u *User) AuthToken() (string, error) {
	c, err := u.claims()
	if err != nil {
		return "", err
	}

	tokenString, err := u.authTokenFromClaims(c)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (u *User) claims() (*UserClaims, error) {
	if u.ID < 1 {
		return nil, errors.New("invalid user id")
	}

	if len(u.Roles) < 1 { // todo:: properly validate roles
		return nil, errors.New("invalid roles")
	}

	//expirationTime := time.Now().Add(5 * time.Minute)
	expirationTime := time.Now().Add(60 * time.Hour * 24 * 3)
	c := &UserClaims{
		UserID: u.ID,
		Roles: u.Roles,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	return c, nil
}

func (u *User) authTokenFromClaims(c *UserClaims) (string, error) {
	jwtKey := []byte("my_secrete_key") // todo:: store somewhere else

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, c)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
