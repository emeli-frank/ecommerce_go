package ecommerce

import (
	"context"
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
	SaveCreditCard(c *CreditCard, custID int) (int, error)
	CreditCards(uid int) ([]CreditCard, error)
	DeleteCreditCard(id int) error
	UpdateCustomerAddress(custID int, a *Address) error
	CustomerAddress(custID int) (*Address, error)
	DeleteCustomerAddress(custID int) error
	OrdersByCustID(custID int) ([]Order, error)
	CartItems(custID int) ([]CartItem, error)
	AddCartItems(custID, productID int) error
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
	AddressID int `json:"address_id"`
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

// context stuff
var userKey key

// NewUserContext returns a new Context that carries value u.
func NewUserContext(ctx context.Context, u *User) context.Context {
	return context.WithValue(ctx, userKey, u)
}

// UserFromContext returns the User value stored in ctx, if any.
func UserFromContext(ctx context.Context) (*User, bool) {
	u, ok := ctx.Value(userKey).(*User)
	return u, ok
}

// UserFromAuthToken returns appropriate user type (admin, org user, or applicant)
// with only minimal info like user id, roles and org id from auth token.
// If more user info is needed, the DB should be queried.
func UserFromAuthToken(authToken string) (*User, error) {
	c := &UserClaims{}

	_, err := jwt.ParseWithClaims(authToken, c, func(token *jwt.Token) (interface{}, error) {
		return []byte("my_secrete_key"), nil
	})

	// todo:: look up, this was in the doc
	/*if claims, ok := token.Claims.(*MyCustomClaims); ok && token.Valid {
		fmt.Printf("%v %v", claims.Foo, claims.StandardClaims.ExpiresAt)
	} else {
		fmt.Println(err)
	}*/

	if err != nil {
		// user is not logged in
		return nil, err
	}

	// user is logged in
	var u2 *User
	u2 = &User{
		ID:    c.UserID,
		Roles: c.Roles,
	}

	return u2, nil
}
