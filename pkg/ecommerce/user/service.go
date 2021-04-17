package user

import (
	"database/sql"
	"ecommerce/pkg/ecommerce"
	errors2 "ecommerce/pkg/ecommerce/errors"
	"errors"
	"golang.org/x/crypto/bcrypt"
)

type repository interface {
	SaveUserWithTx(tx *sql.Tx, user *ecommerce.User, hashedPassword string) (int, error)
	UpdateRolesWithTx(tx *sql.Tx, uid int, roles []int) error
	UserIDAndPasswordByEmail(email string) (int, string, error)
	User(uid int) (*ecommerce.User, error)
	UpdateUserWithTx(tx *sql.Tx, user *ecommerce.User) error
	SaveCreditCard(c *ecommerce.CreditCard, custID int) (int, error)
	CreditCards(uid int) ([]ecommerce.CreditCard, error)
	DeleteCreditCard(id int) error
	//Product(id int) (*ecommerce.Product, error)
	CustOrderIDs(id int) ([]int, error)
	CartItems(custID int) ([]ecommerce.CartItem, error)
	AddCartItems(custID, productID int) error
	Tx() (*sql.Tx, error)
}

type addressRepo interface {
	SaveAddressWithTx(tx *sql.Tx, a *ecommerce.Address) (int, error)
	UpdateAddress(a *ecommerce.Address) error
	Address(id int) (*ecommerce.Address, error)
	DeleteAddress(tx *sql.Tx, id int) error
}

type orderRepo interface {
	SaveOrder(tx *sql.Tx, o *ecommerce.Order) (int, error)
	Orders(ids []int) ([]ecommerce.Order, error)
}

func New(db *sql.DB, repo repository, addressRepo addressRepo, orderRepo orderRepo, productService ecommerce.ProductService) *service {
	return &service{db: db, r: repo, addressRepo: addressRepo, orderRepo: orderRepo, productService: productService}
}

type service struct {
	db *sql.DB
	r repository
	addressRepo addressRepo
	orderRepo orderRepo
	productService ecommerce.ProductService
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

	return u, errors2.Wrap(err, op, "getting user from repo")
}

func (s *service) UpdateUser(user *ecommerce.User) error {
	const op = "userService.UpdateUser"

	tx, err := s.r.Tx()
	if err != nil {
		return errors2.Wrap(err, op, "getting tx")
	}

	err = s.r.UpdateUserWithTx(tx, user)
	if err != nil {
		tx.Rollback()
		return errors2.Wrap(err, op, "updating from repo")
	}

	return errors2.Wrap(tx.Commit(), op, "committing tx")
}

func (s *service) SaveCreditCard(c *ecommerce.CreditCard, custID int) (int, error) {
	const op = "userService.SaveCreditCard"

	// todo:: encrypt card name and number

	id, err := s.r.SaveCreditCard(c, custID)
	return id, errors2.Wrap(err, op, "getting credit card")
}

func (s *service) CreditCards(uid int) ([]ecommerce.CreditCard, error) {
	const op = "userService.CreditCards"

	cc, err := s.r.CreditCards(uid)
	return cc, errors2.Wrap(err, op, "getting credit cards from repo")
}

func (s *service) DeleteCreditCard(id int) error {
	const op = "userService.DeleteCreditCard"

	return errors2.Wrap(s.r.DeleteCreditCard(id), op, "deleting card via repo")
}

func (s *service) UpdateCustomerAddress(custID int, a *ecommerce.Address) error {
	const op = "userService.UpdateCustomerAddress"

	if a.ID > 0 {
		// update address
		// todo:: check that user actually owns address before updating
		return errors2.Wrap(s.addressRepo.UpdateAddress(a), op, "updating address from repo")
	} else {
		// create new address
		u, err := s.r.User(custID)
		if err != nil {
			return errors2.Wrap(err, op, "getting user from repo")
		} else if u.AddressID > 0 {
			return errors2.Wrap(errors.New("can update but not create new address"), op, "checking user address")
		}

		tx, err := s.r.Tx()
		if err != nil {
			return errors2.Wrap(err, op, "getting tx")
		}

		addressID, err := s.addressRepo.SaveAddressWithTx(tx, a)
		if err != nil {
			_ = tx.Rollback()
			return errors2.Wrap(err, op, "saving address from repo")
		}

		u.AddressID = addressID

		// update user with new address
		err = s.r.UpdateUserWithTx(tx, u)
		if err != nil {
			_ = tx.Rollback()
			return errors2.Wrap(err, op, "updating user via repo")
		}

		return errors2.Wrap(tx.Commit(), op, "committing tx")
	}
}

func (s *service) CustomerAddress(custID int) (*ecommerce.Address, error) {
	const op = "userService.CustomerAddress"

	u, err := s.r.User(custID)
	if err != nil {
		return nil, errors2.Wrap(err, op, "getting user form repo")
	} else if u.AddressID < 1 {
		return nil, nil
	}

	a, err := s.addressRepo.Address(u.AddressID)
	return a, errors2.Wrap(err, op, "getting address from repo")
}

func (s *service) DeleteCustomerAddress(custID int) error {
	const op = "userService.DeleteAddress"

	u, err := s.r.User(custID)
	if err != nil {
		return errors2.Wrap(err, op, "getting user form repo")
	}

	tx, err := s.r.Tx()
	if err != nil {
		return errors2.Wrap(err, op, "getting tx")
	}

	// update user
	addressID := u.AddressID // save address id before overwriting
	u.AddressID = 0
	err = s.r.UpdateUserWithTx(tx, u)
	if err != nil {
		tx.Rollback()
		return errors2.Wrap(err, op, "updating user via repo")
	}

	// delete address
	err = s.addressRepo.DeleteAddress(tx, addressID)
	if err != nil {
		tx.Rollback()
		return errors2.Wrap(err, op, "deleting address from repo")
	}

	return errors2.Wrap(tx.Commit(), op, "committing tx")
}

func (s *service) CreateOrder(o *ecommerce.Order) (int, error) {
	const op = "userService.CreateOrder"

	p, err := s.productService.Product(o.Product.ID)
	if err != nil {
		return 0, errors2.Wrap(err, op, "getting product")
	}

	tx, err := s.r.Tx()
	if err != nil {
		return 0, errors2.Wrap(err, op, "obtaining tx")
	}

	p.Quantity = p.Quantity - o.Quantity

	// update product quantity
	err = s.productService.UpdateProductWithTx(tx, p)
	if err != nil {
		tx.Rollback()
		return 0, errors2.Wrap(err, op, "updating product quantity")
	}

	// save order
	orderID, err := s.orderRepo.SaveOrder(tx, o)
	if err != nil {
		tx.Rollback()
		return 0, errors2.Wrap(err, op, "saving order")
	}

	return orderID, errors2.Wrap(tx.Commit(), op, "committing tx")
}

func (s *service) OrdersByCustID(custID int) ([]ecommerce.Order, error) {
	const op = "userService.OrdersByCustID"

	orderIDs, err := s.r.CustOrderIDs(custID)
	if err != nil {
		return nil, errors2.Wrap(err, op, "getting cust order dis")
	}

	oo, err := s.orderRepo.Orders(orderIDs)
	return oo, errors2.Wrap(err, op, "getting orders")
}

func (s *service) CartItems(custID int) ([]ecommerce.CartItem, error) {
	const op = "userService.CartItems"

	// get cart items
	cc, err := s.r.CartItems(custID)
	if err != nil {
		return nil, errors2.Wrap(err, op, "getting cart items from repo")
	}

	var pdtIDs []int
	for _, c := range cc {
		pdtIDs = append(pdtIDs, c.Product.ID)
	}

	// get products with ids
	pp, err := s.productService.ProductsFromIDs(pdtIDs)
	if err != nil {
		return nil, errors2.Wrap(err, op, "getting products from ids")
	}

	// attach products to cart items
	for i := 0; i < len(cc); i++ {
		for _, p := range pp {
			if p.ID == cc[i].Product.ID {
				cc[i].Product = p
			}
		}
	}

	return cc, nil
}

func (s *service) AddCartItems(custID, productID int) error {
	const op = "userService.AddCartItems"

	return errors2.Wrap(s.r.AddCartItems(custID, productID), op, "adding cart item via repo")
}
