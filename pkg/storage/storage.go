package storage

import (
	"database/sql"
	"errors"
	"fmt"

	"reflect"
	"time"
)

var (
	ErrNotAPtr = errors.New("deactivation setter is not a pointer")
)

type Deactivator interface {
	Deactivate(by int, at time.Time)
}

// SetDeactivation receives the pointer of a type that implements deactivationSetter
// and returns an error if type is not a pointer
func SetDeactivation(o Deactivator, deactivatedAt sql.NullTime, deactivatedBy sql.NullInt32) error {
	var at time.Time
	var by int

	if o == nil {
		return errors.New("deactivation setter is nil")
	}

	if reflect.ValueOf(o).Type().Kind() != reflect.Ptr {
		return ErrNotAPtr
	}

	if deactivatedAt.Valid && deactivatedBy.Valid {
		at = deactivatedAt.Time
		by = int(deactivatedBy.Int32)
		o.Deactivate(by, at)
	}

	return nil
}

// todo:: consider moving to storage
type QueryExecutor interface {
	Exec(string, ...interface{}) (sql.Result, error)
}

// todo:: consider moving to storage
type Queryer interface {
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
}

// todo:: move to storage
type DB interface {
	QueryExecutor
	Queryer
}

const (
	MaxBigIntValue = 18446744073709551615
)

func StrToNullableStr(s1 string) sql.NullString {
	var s2 sql.NullString
	if s1 == "" {
		return s2
	}

	s2.Valid = true
	s2.String = s1
	return s2
}

func NullableStrToStr(s1 sql.NullString) string {
	var s2 string
	if !s1.Valid {
		return s2
	}

	return s1.String
}

func IntToNullableInt(n1 int64) sql.NullInt64 {
	var n2 sql.NullInt64
	if n1 == 0 {
		return n2
	}

	n2.Valid = true
	n2.Int64 = n1
	return n2
}

func NullableIntToInt(n1 sql.NullInt64) int64 {
	var n2 int64
	if !n1.Valid {
		return n2
	}

	return n1.Int64
}

func NullableFloatToFloat(n1 sql.NullFloat64) float64 {
	var n2 float64
	if !n1.Valid {
		return n2
	}

	return n1.Float64
}

// IntSliceToCommaSeparatedStr takes a slice of int and returns them as strings
// separated by a comma.
// e.g. []int{1, 2, 3] => 1, 2, 3
func IntSliceToCommaSeparatedStr(ids []int) string {
	str := ""

	for k, id := range ids {
		str += fmt.Sprintf("%d", id)
		if k < (len(ids) - 1) { // if not the last element, add comma separator
			str += ", "
		}
	}

	return str
}
