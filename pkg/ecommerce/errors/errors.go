package errors

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// SetMessage sets public message on error wrapper, this message is to be displayed to the client
func SetMessage(err error, pubMsg string) error {
	// fails silently if error is not of type *wrap
	w, ok := err.(*wrap)
	if ok {
		w.pubMsg = pubMsg
	}
	return err
}

// Message returns public error message to display to client or empty string if error is not wrapped
func Message(err error) string {
	w, ok := err.(*wrap)
	// return empty string if err is not of type *wrap
	if !ok {
		return ""
	}

	if w.pubMsg != "" {
		return w.pubMsg
	} else if w.Err != nil {
		return Message(w.Err)
	} else {
		return ""
	}
}

// WrapWithMsg wraps error and sets public message to be displayed to client
func WrapWithMsg(err error, op string, privMsg string, pubMsg string) error {
	w := Wrap(err, op, privMsg)
	w = SetMessage(w, pubMsg)
	return Wrap(w, op, pubMsg)
}

// Unwrap recursively finds non-wrap error and returns the first one encountered
// (e.g NotFound, Conflict etc...) or returns the error untouched if it is not wrapped
func Unwrap(err error) error {
	u, ok := err.(interface{
		Unwrap() error
	})
	if !ok {
		return err
	}
	return Unwrap(u.Unwrap())
}

// Wrap adds context to already wrapped error or wraps it if it is not.
// If error is nil, nil is returned. This makes it easy to write expressions
// like this: return u, error.Wrap(tx.Commit(), op, "...")
func Wrap(err error, op string, message string) error {
	if err == nil {
		return nil
	}
	return &wrap{Err: err, Op:op, privMsg:message}
}

type wrap struct {
	Err     error
	Op      string
	privMsg string
	pubMsg  string
}

func (c *wrap) Error() string {
	var buf bytes.Buffer

	// Print the current operation in our stack, if any.
	if c.Op != "" {
		_, _ = fmt.Fprintf(&buf, "[%s]: ", c.Op)
	} else {
		_, _ = fmt.Fprint(&buf, "_: ")
	}

	// Print the current additional context in our stack, if any.
	if c.privMsg != "" {
		_, _ = fmt.Fprintf(&buf, "[%s] >> ", c.privMsg)
	} else {
		_, _ = fmt.Fprint(&buf, "_ >> ")
	}

	// If wrapping an error, print its Error() message. Otherwise print the error code & message.
	if c.Err != nil {
		buf.WriteString(c.Err.Error())
	} else {
		_, _ = fmt.Fprintf(&buf, "<Generic error> ")
		buf.WriteString(c.privMsg)
	}

	return buf.String()
}

func(c *wrap) Unwrap() error {
	return c.Err
}

func (c *wrap) MarshalJSON() ([]byte, error) {
	e, ok := Unwrap(c).(json.Marshaler)
	if !ok {
		m := Message(c)
		return json.Marshal(struct {
			//Type string `json:"type"`
			Message string `json:"error,omitempty"`
		}{/*Type:"", */Message:m})
	}

	return e.MarshalJSON()
}

type NotFound struct {
	Err     error
}

// Error outputs stack info that should not be shown to client.
func (e *NotFound) Error() string {
	return e.Err.Error()
}

func (e *NotFound) Cause() error {
	return e.Err
}

