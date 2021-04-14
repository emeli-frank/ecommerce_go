package ecommerce

import (
	"encoding/json"
	"time"
)

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
