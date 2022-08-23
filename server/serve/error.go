package serve

import (
	"encoding/json"
	"net/http"
)

type APIError struct {
	Message string `json:"message"`
	error
}

func NewAPIError(err error) *APIError {
	return &APIError{
		Message: "Issue Dealing with the request",
		error:   err,
	}
}

func (a *APIError) HandleError(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(a)
}
