package http

import (
	"ecommerce/pkg/ecommerce"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
)

type malformedRequest struct {
	status int
	msg    string
}

func (mr *malformedRequest) Error() string {
	return mr.msg
}

func (mr *malformedRequest) MarshalJSON() ([]byte, error) {
	v, err := json.Marshal(struct {
		Type string `json:"type"`
		Error string `json:"errors"`
	}{Type:"validation", Error:mr.msg})

	if err != nil {
		return nil, err
	}

	return v, nil

}

// note:: does not handle error parsing time properly (*time.ParseError)
// this error is seen when trying to encode "" to a time field
func decodeJSONBody(w http.ResponseWriter, r *http.Request, dst interface{}) error {
	if r.Header.Get("Content-Type") != "" && r.Header.Get("Content-Type") != "application/json" {
		msg := "Content-Type header is not application/json"
		return &malformedRequest{status: http.StatusUnsupportedMediaType, msg: msg}
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1048576)

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	err := dec.Decode(&dst)
	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError

		switch {
		case errors.As(err, &syntaxError):
			msg := fmt.Sprintf("Request body contains badly-formed JSON (at position %d)", syntaxError.Offset)
			return &malformedRequest{status: http.StatusBadRequest, msg: msg}

		case errors.Is(err, io.ErrUnexpectedEOF):
			msg := fmt.Sprintf("Request body contains badly-formed JSON")
			return &malformedRequest{status: http.StatusBadRequest, msg: msg}

		case errors.As(err, &unmarshalTypeError):
			msg := fmt.Sprintf("Request body contains an invalid value for the %q field (at position %d)", unmarshalTypeError.Field, unmarshalTypeError.Offset)
			return &malformedRequest{status: http.StatusBadRequest, msg: msg}

		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			msg := fmt.Sprintf("Request body contains unknown field %s", fieldName)
			return &malformedRequest{status: http.StatusBadRequest, msg: msg}

		case errors.Is(err, io.EOF):
			msg := "Request body must not be empty"
			return &malformedRequest{status: http.StatusBadRequest, msg: msg}

		case err.Error() == "http: request body too large":
			msg := "Request body must not be larger than 1MB"
			return &malformedRequest{status: http.StatusRequestEntityTooLarge, msg: msg}

		default:
			return err
		}
	}

	if dec.More() {
		msg := "Request body must only contain a single JSON object"
		return &malformedRequest{status: http.StatusBadRequest, msg: msg}
	}

	return nil
}


type Http struct {
	Response       *response
	ProductService ecommerce.ProductService
	UserService ecommerce.UserService
}

func NewServer(response *response) *Http {
	return &Http{
		Response:    response,
	}
}


// #### AUTHENTICATION ####
func (h Http) authenticate(w http.ResponseWriter, r *http.Request) {

}

func (h Http) createCustomer(w http.ResponseWriter, r *http.Request) {
	data := &struct {
		Customer  ecommerce.Customer `json:"customer"`
		Password string 		  `json:"password"`
	}{}
	if err := decodeJSONBody(w, r, data); err != nil {
		var mr *malformedRequest
		if errors.As(err, &mr) {
			h.Response.clientError(w, mr.status, mr.msg)
		} else {
			h.Response.serverError(w, err)
		}
		return
	}

	id, err := h.UserService.CreateCustomer(&data.Customer, data.Password)
	if err != nil {
		h.Response.serverError(w, err)
		return
	}

	h.Response.respond(w, http.StatusCreated, nil, struct{
		ID int `json:"id"`
	}{ID:id})
}

func (h Http) getProducts(w http.ResponseWriter, r *http.Request) {
	const op = "http.getProducts"

	const size = 20
	var err error
	var page, categoryID int

	if r.FormValue("category") != "" {
		categoryID, err = strconv.Atoi(r.FormValue("category"))
		if err != nil {
			h.Response.clientError(w, http.StatusBadRequest, "category id")
			return
		}
	}

	if r.FormValue("page") != "" {
		page, err = strconv.Atoi(r.FormValue("page"))
		if err != nil {
			h.Response.clientError(w, http.StatusBadRequest, "invalid page number")
			return
		}
	} else {
		page = 1
	}

	var minPrice, maxPrice float32
	var discount int
	if r.FormValue("min-price") != "" {
		p, err := strconv.ParseFloat(r.FormValue("min-price"), 32)
		minPrice = float32(p)
		if err != nil {
			h.Response.clientError(w, http.StatusBadRequest, "invalid min price")
			return
		}
	}

	if r.FormValue("max-price") != "" {
		p, err := strconv.ParseFloat(r.FormValue("max-price"), 32)
		maxPrice = float32(p)
		if err != nil {
			h.Response.clientError(w, http.StatusBadRequest, "invalid max price")
			return
		}
	}

	if r.FormValue("discount") != "" {
		discount, err = strconv.Atoi(r.FormValue("discount"))
		if err != nil {
			h.Response.clientError(w, http.StatusBadRequest, "invalid discount")
			return
		}
	}

	filter := &ecommerce.ProductFilter{
		MinPrice: minPrice,
		MaxPrice: maxPrice,
		Discount: discount,
	}

	pp, err := h.ProductService.Products(categoryID, r.FormValue("q"), filter, page, size)
	if err != nil {
		h.Response.serverError(w, err)
		return
	}

	if pp == nil {
		pp = []ecommerce.Product{}
	}

	h.Response.respond(w, http.StatusOK, nil, pp)
}
