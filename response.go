package responce

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// Response is the common JSON envelope written by this package.
type Response struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
	Errors  any    `json:"errors,omitempty"`
	Meta    any    `json:"meta,omitempty"`
}

// ValidationErrors maps field names to one or more validation messages.
type ValidationErrors map[string][]string

// WriteJSON marshals payload before writing headers, so marshal errors are
// returned before the response is committed.
func WriteJSON(w http.ResponseWriter, statusCode int, payload any) error {
	if statusCode < 100 || statusCode > 599 {
		return fmt.Errorf("invalid HTTP status code %d", statusCode)
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal JSON response: %w", err)
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)
	if _, err := w.Write(body); err != nil {
		return fmt.Errorf("write JSON response: %w", err)
	}
	return nil
}

// WriteSuccess writes a successful response with the supplied status code.
func WriteSuccess(w http.ResponseWriter, statusCode int, message string, data any) error {
	return WriteJSON(w, statusCode, Response{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// WriteError writes an error response with the supplied status code.
func WriteError(w http.ResponseWriter, statusCode int, message string) error {
	return WriteJSON(w, statusCode, Response{
		Success: false,
		Message: message,
	})
}

// WriteValidationError writes validation details with HTTP 422 Unprocessable Entity.
func WriteValidationError(w http.ResponseWriter, message string, errors ValidationErrors) error {
	return WriteJSON(w, http.StatusUnprocessableEntity, Response{
		Success: false,
		Message: message,
		Errors:  errors,
	})
}

// WritePaginated writes a response with pagination metadata.
func WritePaginated(
	w http.ResponseWriter,
	statusCode int,
	message string,
	data any,
	page int,
	perPage int,
	totalItems int64,
) error {
	meta, err := NewPaginationMeta(page, perPage, totalItems)
	if err != nil {
		return err
	}

	return WriteJSON(w, statusCode, Response{
		Success: true,
		Message: message,
		Data:    data,
		Meta:    meta,
	})
}
