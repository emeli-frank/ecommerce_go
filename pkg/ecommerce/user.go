package ecommerce

const (
	RoleCustomer = 1
)
type UserService interface {
	CreateCustomer(c *Customer, password string) (int, error)
}

type User interface {
	UserID() int
}

// UserBase base user struct to be embedded in different user types.
type UserBase struct {
	ID int `json:"id"`
	FirstName string `json:"first_name"`
	LastName string `json:"last_name"`
	Email string `json:"email"`
	Roles []int `json:"roles"`
}

func (u *UserBase) UserID() int { return u.ID }

type Customer struct {
	UserBase
}
