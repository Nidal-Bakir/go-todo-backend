package server

import (
	"encoding/json"
	"net/http"
)

func WriteJson(w http.ResponseWriter, code int, payload any) {
	w.Header().Add("Content-Type", "application/json")

	bytes, err := json.Marshal(payload)
	if err == nil {
		w.WriteHeader(code)
		w.Write(bytes)
		return
	}

	w.WriteHeader(http.StatusInternalServerError)
	
	type ErrorRes struct {
		ErrorStr string `json:"error"`
	}

	bytes, err = json.Marshal(ErrorRes{ErrorStr: err.Error()})
	if err == nil {
		goto write
	}

	bytes, err = json.Marshal(ErrorRes{ErrorStr: err.Error()})
	if err == nil {
		goto write
	}

	bytes, err = json.Marshal(ErrorRes{ErrorStr: "Unknown Error while marshaling error struct in write json func"})
	if err == nil {
		goto write
	}

	bytes = []byte("We should not rich this line something really bad is happening!!")

write:
	w.Write(bytes)
}
