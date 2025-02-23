package utils

import (
	"encoding/json"
	"net/http"
)

type JSONResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func SendJSONResponse(w http.ResponseWriter, status int, res JSONResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	json.NewEncoder(w).Encode(res)
}
