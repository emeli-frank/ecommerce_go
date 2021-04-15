package main

import (
	"ecommerce/pkg/ecommerce/product"
	"ecommerce/pkg/ecommerce/user"
	http2 "ecommerce/pkg/http"
	"ecommerce/pkg/storage"
	"ecommerce/pkg/storage/postgres"
	"flag"
	"fmt"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	"os"
)

func main() {
	addr := flag.String("addr", ":5000", "HTTP network address")
	dsn := flag.String("dsn", "host=localhost port=5432 user=ecommerce password=password dbname=ecommerce sslmode=disable", "Postgresql database connection info")
	flag.Parse()

	//infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	db, err := storage.OpenDB("postgres", *dsn)
	if err != nil {
		errorLog.Fatal(err)
	}
	defer db.Close()

	response := http2.NewResponse(errorLog)

	productRepo := postgres.NewProductStorage(db)
	productService := product.New(db, productRepo)

	userRepo := postgres.NewUserStorage(db)
	userService := user.New(db, userRepo)

	httpEndpoint := &http2.Http{
		Response: response,
		ProductService: productService,
		UserService: userService,
	}
	router := httpEndpoint.Routes()

	srv := &http.Server{
		Addr: *addr,
		Handler: router,
		ErrorLog: errorLog,
	}

	fmt.Printf("Starting server on: %s\n", *addr)
	err = srv.ListenAndServe()
	errorLog.Fatal(err)
}

