# ecommerce_go
A simple eCommerce REST application written in Go.
## Installation

### Compile the application
- Install go ([https://https://golang.org/doc/install](https://https://golang.org/doc/install))
- Navigate to the project directory and run `go build`. This generates an executable file in that same directory.


### Set up the database
- Install postgresql
- Set up a user and database for the applicaton like so
`CREATE DATABASE "ecommerce";`
`CREATE USER ecommerce WITH ENCRYPTED PASSWORD 'password';`
`GRANT ALL PRIVILEGES ON DATABASE ecommerce TO ecommerce;`

### Running the application
#### Set up mock data
- Run `go run cmd/cli/main.go`. If propted, selected the second option to seed db with mock data.
- Start the compiled application in your command terminal
