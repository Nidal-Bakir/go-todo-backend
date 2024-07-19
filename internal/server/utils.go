package server

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type ErrorRes struct {
	Error  error   `json:"error"`
	Errors []error `json:"errors,omitempty"`
}

func (e ErrorRes) MarshalJSON() ([]byte, error) {
	m := make(map[string]any)
	m["error"] = e.Error.Error()

	errsLen := len(e.Errors)
	if errsLen != 0 {
		errors := make([]string, errsLen)
		for i, e := range e.Errors {
			errors[i] = e.Error()
		}
		m["errors"] = errors
	}

	return json.Marshal(m)
}

func WriteJson(w http.ResponseWriter, code int, payload any) {
	w.Header().Add("Content-Type", "application/json")

	bytes, err := json.Marshal(payload)
	if err == nil {
		w.WriteHeader(code)
		w.Write(bytes)
		return
	}

	w.WriteHeader(http.StatusInternalServerError)

	bytes, err = json.Marshal(ErrorRes{Error: err})
	if err == nil {
		goto write
	}

	bytes, err = json.Marshal(ErrorRes{Error: err})
	if err == nil {
		goto write
	}

	bytes, err = json.Marshal(ErrorRes{Error: fmt.Errorf("unknown Error while marshaling error struct in write json func")})
	if err == nil {
		goto write
	}

	bytes = []byte("We should not rich this line something really bad is happening!!")

write:
	w.Write(bytes)
}
