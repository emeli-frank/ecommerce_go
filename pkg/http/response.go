package http

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"runtime/debug"
)

type response struct {
	errorLog *log.Logger
}

func NewResponse(errorLog *log.Logger) *response {
	return &response{
		errorLog: errorLog,
	}
}

func (r response) respond(
	w http.ResponseWriter,
	statusCode int,
	headers responseHeaders,
	output interface{},
) {
	var contentTypeFound bool
	if headers != nil {
		for _, h := range headers {
			for k, v := range h {
				if k == "Content-Type" {
					contentTypeFound = true
				}
				w.Header().Set(k, v)
			}
		}
	}

	w.Header().Set("X-Content-Type-Options", "nosniff")
	if !contentTypeFound {
		w.Header().Set("Content-Type", "application/json")
	}
	w.WriteHeader(statusCode)

	if output == nil {
		return
	}

	err := json.NewEncoder(w).Encode(output)
	if err != nil {
		// todo:: see if it should be handled
		// the err var was not there before
		fmt.Println(err)
	}
}

type ErrorMessage struct {
	Message interface{} `json:"message"`
}

type ErrorOutput struct {
	Error ErrorMessage `json:"error"`
}

func (e *ErrorOutput) MarshalJSON() ([]byte, error) {
	return json.Marshal(e)
}

func (r response) clientError(w http.ResponseWriter, code int, message interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(code)

	var output interface{}

	if message == nil {
		return
	}

	switch m := message.(type) {
	case string:
		if message == "" {
			return
		}

		output = ErrorOutput{
			Error: ErrorMessage{Message:m},
		}
	case json.Marshaler:
		output = ErrorOutput{
			Error: ErrorMessage{Message:m},
		}
	}

	o, _ := json.Marshal(output)
	_, _ = fmt.Fprintln(w, string(o))
}

func (r response) serverError(w http.ResponseWriter, err error) {
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
	_ = r.errorLog.Output(2, trace)
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(http.StatusInternalServerError)
}

type responseHeaders []map[string]string

