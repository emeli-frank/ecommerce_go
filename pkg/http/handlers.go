package http

import (
	"ecommerce/pkg/ecommerce"
	errors2 "ecommerce/pkg/ecommerce/errors"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
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

var ErrUserNotFoundInRequestCtx = errors.New("user not found in request context")

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
	const op = "http.authenticate"

	data := &struct {
		Email string `json:"email"`
		Password string `json:"password"`
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

	// check if email and password match
	match, uid, err := h.UserService.EmailMatchPassword(data.Email, data.Password)
	if err != nil {
		h.Response.serverError(w, err)
		return
	} else if !match {
		h.Response.clientError(w, http.StatusUnauthorized, "email and password didn't match")
		return
	}

	// get user
	u, err := h.UserService.User(uid)
	if err != nil {
		h.Response.serverError(w, err)
		return
	}

	// get auth token
	authToken, err := u.AuthToken()
	if err != nil {
		h.Response.serverError(w, err)
		return
	}

	o := struct {
		AuthorizationToken string `json:"auth_token,omitempty"`
		User interface{} `json:"user"`
	}{
		AuthorizationToken:         authToken,
		User:                       u,
	}

	h.Response.respond(w, http.StatusOK, nil, o)
}

func (h Http) createCustomer(w http.ResponseWriter, r *http.Request) {
	data := &struct {
		Customer  ecommerce.User `json:"customer"`
		Password string          `json:"password"`
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

func (h Http) updateCustomer(w http.ResponseWriter, r *http.Request) {
	var u ecommerce.User
	if err := decodeJSONBody(w, r, &u); err != nil {
		var mr *malformedRequest
		if errors.As(err, &mr) {
			h.Response.clientError(w, mr.status, mr.msg)
		} else {
			h.Response.serverError(w, err)
		}
		return
	}

	err := h.UserService.UpdateUser(&u)
	if err != nil {
		h.Response.serverError(w, err)
		return
	}

	h.Response.respond(w, http.StatusOK, nil, nil)
}

func (h Http) getCustomerAddress(w http.ResponseWriter, r *http.Request) {
	// get user from request context
	u, ok := ecommerce.UserFromContext(r.Context())
	if !ok {
		h.Response.serverError(w, ErrUserNotFoundInRequestCtx)
		return
	}

	a, err := h.UserService.CustomerAddress(u.ID)
	if err != nil {
		_, ok := errors2.Unwrap(err).(*errors2.NotFound)
		if ok {
			h.Response.respond(w, http.StatusOK, nil, nil)
			return
		}

		h.Response.serverError(w, err)
		return
	}

	h.Response.respond(w, http.StatusOK, nil, a)
}

func (h Http) updateCustomerAddress(w http.ResponseWriter, r *http.Request) {
	var a ecommerce.Address
	if err := decodeJSONBody(w, r, &a); err != nil {
		var mr *malformedRequest
		if errors.As(err, &mr) {
			h.Response.clientError(w, mr.status, mr.msg)
		} else {
			h.Response.serverError(w, err)
		}
		return
	}

	// get user from request context
	u, ok := ecommerce.UserFromContext(r.Context())
	if !ok {
		h.Response.serverError(w, ErrUserNotFoundInRequestCtx)
		return
	}

	err := h.UserService.UpdateCustomerAddress(u.ID, &a)
	if err != nil {
		h.Response.serverError(w, err)
		return
	}

	h.Response.respond(w, http.StatusOK, nil, nil)
}

func (h Http) deleteCustomerAddress(w http.ResponseWriter, r *http.Request) {
	// get user from request context
	u, ok := ecommerce.UserFromContext(r.Context())
	if !ok {
		h.Response.serverError(w, ErrUserNotFoundInRequestCtx)
		return
	}

	err := h.UserService.DeleteCustomerAddress(u.ID)
	if err != nil {
		h.Response.serverError(w, err)
		return
	}

	h.Response.respond(w, http.StatusOK, nil, nil)
}

func (h Http) saveCreditCard(w http.ResponseWriter, r *http.Request) {
	var c *ecommerce.CreditCard

	if err := decodeJSONBody(w, r, &c); err != nil {
		var mr *malformedRequest
		if errors.As(err, &mr) {
			h.Response.clientError(w, mr.status, mr.msg)
		} else {
			h.Response.serverError(w, err)
		}
		return
	}

	// get user from request context
	u, ok := ecommerce.UserFromContext(r.Context())
	if !ok {
		h.Response.serverError(w, ErrUserNotFoundInRequestCtx)
		return
	}

	id, err := h.UserService.SaveCreditCard(c, u.ID)
	if err != nil {
		h.Response.serverError(w, err)
		return
	}

	h.Response.respond(w, http.StatusCreated, nil, struct{
		ID int `json:"id"`
	}{ID:id})
}

func (h Http) deleteCreditCard(w http.ResponseWriter, r *http.Request) {
	cardID, err := strconv.Atoi(mux.Vars(r)["cardID"])
	if err != nil {
		h.Response.clientError(w, http.StatusBadRequest, "invalid card id")
		return
	}

	err = h.UserService.DeleteCreditCard(cardID)
	if err != nil {
		h.Response.serverError(w, err)
		return
	}

	h.Response.respond(w, http.StatusOK, nil, nil)
}

func (h Http) getCreditCard(w http.ResponseWriter, r *http.Request) {

	// get user from request context
	u, ok := ecommerce.UserFromContext(r.Context())
	if !ok {
		h.Response.serverError(w, ErrUserNotFoundInRequestCtx)
		return
	}

	cc, err := h.UserService.CreditCards(u.ID)
	if err != nil {
		h.Response.serverError(w, err)
		return
	}

	if cc == nil { cc = []ecommerce.CreditCard{} }

	h.Response.respond(w, http.StatusOK, nil, cc)
}

func (h Http) getCustomerOrders(w http.ResponseWriter, r *http.Request) {

	// get user from request context
	u, ok := ecommerce.UserFromContext(r.Context())
	if !ok {
		h.Response.serverError(w, ErrUserNotFoundInRequestCtx)
		return
	}

	oo, err := h.UserService.OrdersByCustID(u.ID)
	if err != nil {
		h.Response.serverError(w, err)
		return
	}

	if oo == nil { oo = []ecommerce.Order{} }

	h.Response.respond(w, http.StatusOK, nil, oo)
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

func (h Http) getProduct(w http.ResponseWriter, r *http.Request) {
	const op = "http.getProduct"

	pdtID, err := strconv.Atoi(mux.Vars(r)["productID"])
	if err != nil {
		h.Response.clientError(w, http.StatusBadRequest, "invalid product id")
		return
	}

	p, err := h.ProductService.Product(pdtID)
	if err != nil {
		h.Response.serverError(w, err)
		return
	}

	h.Response.respond(w, http.StatusOK, nil, p)
}

func (h Http) getCartItems(w http.ResponseWriter, r *http.Request) {

	// get user from request context
	u, ok := ecommerce.UserFromContext(r.Context())
	if !ok {
		h.Response.serverError(w, ErrUserNotFoundInRequestCtx)
		return
	}

	cc, err := h.UserService.CartItems(u.ID)
	if err != nil {
		h.Response.serverError(w, err)
		return
	}

	if cc == nil { cc = []ecommerce.CartItem{} }

	h.Response.respond(w, http.StatusOK, nil, cc)
}

func (h Http) addCartItems(w http.ResponseWriter, r *http.Request) {

	var data struct {
		ProductID int `json:"product_id"`
	}
	if err := decodeJSONBody(w, r, &data); err != nil {
		var mr *malformedRequest
		if errors.As(err, &mr) {
			h.Response.clientError(w, mr.status, mr.msg)
		} else {
			h.Response.serverError(w, err)
		}
		return
	}

	// get user from request context
	u, ok := ecommerce.UserFromContext(r.Context())
	if !ok {
		h.Response.serverError(w, ErrUserNotFoundInRequestCtx)
		return
	}

	err := h.UserService.AddCartItems(u.ID, data.ProductID)
	if err != nil {
		h.Response.serverError(w, err)
		return
	}

	h.Response.respond(w, http.StatusOK, nil, nil)
}
