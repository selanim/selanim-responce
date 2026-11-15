package responce

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
)

// Logger is the logging interface used by this package.
type Logger interface {
	Printf(format string, v ...any)
}

var (
	loggerMu sync.RWMutex
	logger   Logger = log.Default()
)

// SetLogger configures package logging. Passing nil disables logging.
// The logger should be safe for concurrent use.
func SetLogger(l Logger) {
	loggerMu.Lock()
	logger = l
	loggerMu.Unlock()
}

func logf(format string, v ...any) {
	loggerMu.RLock()
	l := logger
	loggerMu.RUnlock()
	if l != nil {
		l.Printf(format, v...)
	}
}

// Response is the common JSON envelope written by this package.
type Response struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Code    string `json:"code,omitempty"`
	Data    any    `json:"data,omitempty"`
	Errors  any    `json:"errors,omitempty"`
	Meta    any    `json:"meta,omitempty"`
}

// ValidationErrors maps field names to one or more validation messages.
type ValidationErrors map[string][]string

// WriteJSON marshals payload before writing headers, so marshal errors are
// returned before the response is committed.
func WriteJSON(w http.ResponseWriter, statusCode int, payload any) error {
	return WriteJSONWithHeaders(w, statusCode, payload, nil)
}

// WriteJSONWithHeaders writes a JSON response with additional headers.
// Additional headers replace existing values with the same name; Content-Type
// is always set to the JSON content type.
func WriteJSONWithHeaders(w http.ResponseWriter, statusCode int, payload any, headers http.Header) error {
	return writeJSONWithContentType(w, statusCode, payload, headers, "application/json; charset=utf-8")
}

func writeJSONWithContentType(
	w http.ResponseWriter,
	statusCode int,
	payload any,
	headers http.Header,
	contentType string,
) error {
	if statusCode < 100 || statusCode > 599 {
		err := fmt.Errorf("invalid HTTP status code %d", statusCode)
		logf("responce: %v", err)
		return err
	}

	body, err := json.Marshal(payload)
	if err != nil {
		err = fmt.Errorf("marshal JSON response: %w", err)
		logf("responce: %v", err)
		return err
	}

	for key, values := range headers {
		canonicalKey := http.CanonicalHeaderKey(key)
		if canonicalKey == "Content-Type" {
			continue
		}
		w.Header()[canonicalKey] = append([]string(nil), values...)
	}
	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(statusCode)
	if _, err := w.Write(body); err != nil {
		err = fmt.Errorf("write JSON response: %w", err)
		logf("responce: %v", err)
		return err
	}
	logf("responce: wrote HTTP response with status %d", statusCode)
	return nil
}

// WriteNoContent writes an HTTP 204 No Content response without a body.
func WriteNoContent(w http.ResponseWriter) error {
	w.WriteHeader(http.StatusNoContent)
	logf("responce: wrote HTTP response with status %d", http.StatusNoContent)
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
	return WriteErrorWithCode(w, statusCode, "", message)
}

// WriteErrorWithCode writes an error response with a machine-readable code.
func WriteErrorWithCode(w http.ResponseWriter, statusCode int, code, message string) error {
	return WriteJSON(w, statusCode, Response{
		Success: false,
		Message: message,
		Code:    code,
	})
}

// WriteValidationError writes validation details with HTTP 422 Unprocessable Entity.
func WriteValidationError(w http.ResponseWriter, message string, errors ValidationErrors) error {
	return WriteValidationErrorWithCode(w, "", message, errors)
}

// WriteValidationErrorWithCode writes validation details with a machine-readable code.
func WriteValidationErrorWithCode(w http.ResponseWriter, code, message string, errors ValidationErrors) error {
	return WriteJSON(w, http.StatusUnprocessableEntity, Response{
		Success: false,
		Message: message,
		Code:    code,
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
		logf("responce: create paginated response: %v", err)
		return err
	}

	return WriteJSON(w, statusCode, Response{
		Success: true,
		Message: message,
		Data:    data,
		Meta:    meta,
	})
}

// WriteCursorPaginated writes a response with opaque cursor pagination metadata.
// The caller is responsible for generating stable cursors from its data source.
func WriteCursorPaginated(
	w http.ResponseWriter,
	statusCode int,
	message string,
	data any,
	nextCursor string,
	previousCursor string,
) error {
	return WriteJSON(w, statusCode, Response{
		Success: true,
		Message: message,
		Data:    data,
		Meta: CursorPaginationMeta{
			NextCursor:     nextCursor,
			PreviousCursor: previousCursor,
			HasNext:        nextCursor != "",
			HasPrevious:    previousCursor != "",
		},
	})
}
