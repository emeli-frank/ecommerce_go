package http

import (
	"github.com/gorilla/mux"
	"github.com/justinas/alice"
	"github.com/rs/cors"
	"net/http"
)

func (h Http) Routes() http.Handler {
	standardMiddleWare := alice.New(h.recoverPanic , h.setReqCtxUser)
	//authOnlyMiddleWare := alice.New(/*s.checkJWT, */s.authenticatedOnly)

	r := mux.NewRouter()

	r.Handle("/customers", http.HandlerFunc(h.createCustomer)).Methods("POST")

	r.Handle("/customers/cards", http.HandlerFunc(h.saveCreditCard)).Methods("POST")

	r.Handle("/customers/cards/{cardID:[0-9]+}", http.HandlerFunc(h.deleteCreditCard)).Methods("DELETE")

	r.Handle("/customers/{uid:[0-9]+}/address", http.HandlerFunc(h.updateCustomerAddress)).Methods("PUT")

	r.Handle("/customers/{uid:[0-9]+}/address", http.HandlerFunc(h.deleteCustomerAddress)).Methods("DELETE")

	r.Handle("/customers/{uid:[0-9]+}/address", http.HandlerFunc(h.getCustomerAddress))

	r.Handle("/customers/{uid:[0-9]+}/cart", http.HandlerFunc(h.addCartItems)).Methods("POST")

	r.Handle("/customers/{uid:[0-9]+}/cart", http.HandlerFunc(h.getCartItems))

	r.Handle("/customers/{uid:[0-9]+}/cart/count", http.HandlerFunc(h.cartItemCount))

	r.Handle("/customers/{uid:[0-9]+}/orders", http.HandlerFunc(h.getCustomerOrders))

	r.Handle("/customers/cards", http.HandlerFunc(h.getCreditCard))

	r.Handle("/users/{uid:[0-9]+}", http.HandlerFunc(h.updateCustomer)).Methods("PUT")

	r.Handle("/users/authentication", http.HandlerFunc(h.authenticate)).Methods("POST")

	r.Handle("/products", http.HandlerFunc(h.getProducts))

	r.Handle("/products/{productID:[0-9]+}", http.HandlerFunc(h.getProduct))

	c := cors.New(cors.Options{
		AllowedOrigins: []string{"http://localhost:4200", "*"}, // todo:: adjust before production
		AllowedMethods: []string{"GET", "POST", "DELETE", "PUT"},
		//AllowedHeaders: []string{"Authorization", "User-Agent", "Sec-Fetch-Dest", "Referer", "Content-Type", "Accept"},
		AllowedHeaders: []string{"*"},
	})
	return c.Handler(standardMiddleWare.Then(r))
	//return cors.Default().Handler(globalMiddleware.Then(r))
}
