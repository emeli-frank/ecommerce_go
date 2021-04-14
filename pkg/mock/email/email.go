package email

import (
	"ecommerce/pkg/ecommerce"
	"fmt"
	"io/ioutil"
	"os"
	"time"
)

func New() ecommerce.Email {
	return &email{}
}

type email struct {}

func (e *email) Send(msg string) error {
	// make directory if not exist
	err := os.MkdirAll("/tmp/emails", os.ModePerm)
	if err != nil {
		return err
	}

	// write file
	err = ioutil.WriteFile(fmt.Sprintf("/tmp/emails/email-%d.html", time.Now().UnixNano()), []byte(msg), 0644)
	if err != nil {
		return err
	}

	return nil
}
