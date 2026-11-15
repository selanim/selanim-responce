package responce

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
)

func ExampleRequestIDMiddleware() {
	SetLogger(log.New(io.Discard, "", 0))
	defer SetLogger(log.Default())

	handler := RequestIDMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := RequestIDFromContext(r.Context())
		_ = WriteSuccess(w, http.StatusOK, "imefanikiwa", map[string]string{
			"request_id": requestID,
		})
	}))

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))

	fmt.Println(recorder.Code == http.StatusOK)
	fmt.Println(len(recorder.Header().Get(RequestIDHeader)) == 32)
	// Output:
	// true
	// true
}

func ExampleWritePaginated() {
	SetLogger(log.New(io.Discard, "", 0))
	defer SetLogger(log.Default())

	recorder := httptest.NewRecorder()
	err := WritePaginated(
		recorder,
		http.StatusOK,
		"Matokeo yamepatikana",
		[]string{"Amina", "Juma"},
		1,
		2,
		3,
	)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(recorder.Code)
	fmt.Println(recorder.Body.String())
	// Output:
	// 200
	// {"success":true,"message":"Matokeo yamepatikana","data":["Amina","Juma"],"meta":{"page":1,"per_page":2,"total_items":3,"total_pages":2,"has_next":true,"has_previous":false}}
}
