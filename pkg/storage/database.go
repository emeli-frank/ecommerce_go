package storage

import (
	"database/sql"
	//_ "github.com/go-sql-driver/mysql"
	"io/ioutil"
	"strings"
)

/*func DropTable(db *sql.DB, name string) error {
	stmt := "DROP TABLE IF EXISTS ?"
	_, err := db.Exec(stmt, name)
	if err != nil {
		return err
	}
	return nil
}*/

func OpenDB(driverName, dsn string) (*sql.DB, error) {
	db, err := sql.Open(driverName, dsn)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

// ExecScripts will receive slice of paths (string) and execute all sql
// statements in it in the order in which they are passed
func ExecScripts(db *sql.DB, paths ...string) error {
	for _, p := range paths {
		script, err := ioutil.ReadFile(p)
		if err != nil {
			return err
		}

		if strings.TrimSpace(string(script)) == "" {
			break
		}

		_, err = db.Exec(string(script))
		if err != nil {
			return err
		}
	}

	return nil
}
