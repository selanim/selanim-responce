package responce

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
)

// RequestIDHeader is the response header set by RequestIDMiddleware.
const RequestIDHeader = "X-Request-ID"

type requestIDContextKey struct{}

// RequestIDFromContext returns the request ID added by RequestIDMiddleware.
func RequestIDFromContext(ctx context.Context) string {
	requestID, _ := ctx.Value(requestIDContextKey{}).(string)
	return requestID
}

// RequestIDMiddleware generates a new ID for every request, puts it in the
// X-Request-ID response header, and adds it to the request context.
func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID, err := newRequestID()
		if err != nil {
			logf("responce: generate request ID: %v", err)
			http.Error(w, "could not generate request ID", http.StatusInternalServerError)
			return
		}

		w.Header().Set(RequestIDHeader, requestID)
		ctx := context.WithValue(r.Context(), requestIDContextKey{}, requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func newRequestID() (string, error) {
	var id [16]byte
	if _, err := rand.Read(id[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(id[:]), nil
}
