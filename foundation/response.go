package foundation

import (
	"encoding/json"
	"net/http"
)

// Response is the response structure for error messages
// and success messages
type Response struct {
	Status    int    `json:"status"`
	Message   string `json:"message"`
	ErrorCode string `json:"error-code"`
}

// ErrorResponse give details to the user about the error that occurred
func ErrorResponse(w http.ResponseWriter, code int, errr error, customCode string) {
	w.Header().Set("Content-Type", "application/json")
	_ = JSONResponse(w, code, errr.Error(), customCode)
}

// JSONResponse converts response into a json
func JSONResponse(w http.ResponseWriter, status int, message, errCode string) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	res := NewResponse(status, message, errCode)
	err := json.NewEncoder(w).Encode(res)
	if err != nil {
		return err
	}
	return nil
}

// NewResponse creates an instance of the response structure
func NewResponse(status int, message, errCode string) *Response {
	return &Response{
		Status:    status,
		Message:   message,
		ErrorCode: errCode,
	}
}
