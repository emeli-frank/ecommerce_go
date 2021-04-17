package ecommerce

import (
	"encoding/json"
	"time"
)

type key int

type Time struct {
	time.Time
}

func (t Time) MarshalJSON() ([]byte, error) {
	if t.IsZero() {
		return []byte("null"), nil
	}

	return json.Marshal(t)
}

type Email interface {
	Send(msg string) error
}

type Price struct {
	Current float32 `json:"current"`
	Old float32 `json:"old,omitempty"`
}

type ProductFilter struct {
	MinPrice float32 `json:"min_price"`
	MaxPrice float32 `json:"max_price"`
	Discount int `json:"discount"`
}

type CreditCard struct {
	ID int `json:"id"`
	Name string `json:"name"`
	Number string `json:"number"`
	ExpiryDate string `json:"expiry_date"`
	CVC string`json:"cvc"`
}

type Address struct {
	ID int `json:"id"`
	Country string `json:"country"`
	State string `json:"state"`
	City string `json:"city"`
	PostalCode string `json:"postal_code"`
	Address string `json:"address"`
}
