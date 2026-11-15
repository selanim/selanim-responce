package responce

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

func TestWriteJSONLogsStatusAndErrors(t *testing.T) {
	var logs bytes.Buffer
	SetLogger(log.New(&logs, "", 0))
	defer SetLogger(log.Default())

	if err := WriteJSON(httptest.NewRecorder(), http.StatusOK, map[string]string{"message": "ok"}); err != nil {
		t.Fatalf("WriteJSON() error = %v", err)
	}
	if !strings.Contains(logs.String(), "wrote HTTP response with status 200") {
		t.Fatalf("success log = %q", logs.String())
	}

	logs.Reset()
	if err := WriteJSON(httptest.NewRecorder(), http.StatusOK, make(chan int)); err == nil {
		t.Fatal("WriteJSON() error = nil, want marshal error")
	}
	if !strings.Contains(logs.String(), "marshal JSON response") {
		t.Fatalf("error log = %q", logs.String())
	}
}

func TestSetLoggerNilDisablesLogging(t *testing.T) {
	var logs bytes.Buffer
	SetLogger(log.New(&logs, "", 0))
	defer SetLogger(log.Default())

	SetLogger(nil)
	if err := WriteJSON(httptest.NewRecorder(), http.StatusOK, nil); err != nil {
		t.Fatalf("WriteJSON() error = %v", err)
	}
	if logs.Len() != 0 {
		t.Fatalf("logs = %q, want no logs", logs.String())
	}
}

func TestWriteSuccess(t *testing.T) {
	recorder := httptest.NewRecorder()
	data := map[string]string{"id": "42"}

	if err := WriteSuccess(recorder, http.StatusCreated, "Created", data); err != nil {
		t.Fatalf("WriteSuccess() error = %v", err)
	}
	if recorder.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusCreated)
	}
	if got := recorder.Header().Get("Content-Type"); got != "application/json; charset=utf-8" {
		t.Fatalf("Content-Type = %q", got)
	}

	var got Response
	if err := json.Unmarshal(recorder.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !got.Success || got.Message != "Created" {
		t.Fatalf("response = %+v", got)
	}
}

func TestWriteNoContent(t *testing.T) {
	recorder := httptest.NewRecorder()

	if err := WriteNoContent(recorder); err != nil {
		t.Fatalf("WriteNoContent() error = %v", err)
	}
	if recorder.Code != http.StatusNoContent || recorder.Body.Len() != 0 {
		t.Fatalf("status/body = %d/%q, want 204/empty", recorder.Code, recorder.Body.String())
	}
	if got := recorder.Header().Get("Content-Type"); got != "" {
		t.Fatalf("Content-Type = %q, want empty", got)
	}
}

func TestWriteErrorWithCode(t *testing.T) {
	recorder := httptest.NewRecorder()

	if err := WriteErrorWithCode(recorder, http.StatusNotFound, "USER_NOT_FOUND", "User hapatikani"); err != nil {
		t.Fatalf("WriteErrorWithCode() error = %v", err)
	}

	var got Response
	if err := json.Unmarshal(recorder.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if got.Success || got.Code != "USER_NOT_FOUND" {
		t.Fatalf("response = %+v", got)
	}
}

func TestWriteProblem(t *testing.T) {
	recorder := httptest.NewRecorder()
	problem := ProblemDetails{
		Type:   "https://example.com/problems/not-found",
		Title:  "Not found",
		Detail: "The requested user was not found.",
		Extensions: map[string]any{
			"code": "USER_NOT_FOUND",
		},
	}

	if err := WriteProblem(recorder, http.StatusNotFound, problem); err != nil {
		t.Fatalf("WriteProblem() error = %v", err)
	}
	if got := recorder.Header().Get("Content-Type"); got != "application/problem+json" {
		t.Fatalf("Content-Type = %q", got)
	}
	var got map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode problem: %v", err)
	}
	if got["type"] != problem.Type || got["title"] != problem.Title ||
		got["status"] != float64(http.StatusNotFound) || got["code"] != "USER_NOT_FOUND" {
		t.Fatalf("problem = %+v", got)
	}
}

func TestWriteProblemRejectsConflictingStatusAndExtensions(t *testing.T) {
	tests := []struct {
		name    string
		problem ProblemDetails
	}{
		{
			name:    "status mismatch",
			problem: ProblemDetails{Status: http.StatusBadRequest},
		},
		{
			name: "extension conflicts",
			problem: ProblemDetails{
				Extensions: map[string]any{"status": "override"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			if err := WriteProblem(recorder, http.StatusNotFound, tt.problem); err == nil {
				t.Fatal("WriteProblem() error = nil, want error")
			}
			if recorder.Code != http.StatusOK || recorder.Body.Len() != 0 {
				t.Fatalf("response committed unexpectedly: status=%d body=%q", recorder.Code, recorder.Body.String())
			}
		})
	}
}

func TestWriteJSONWithHeaders(t *testing.T) {
	recorder := httptest.NewRecorder()
	headers := http.Header{
		"X-Request-ID":  {"request-123"},
		"Cache-Control": {"no-store"},
		"Content-Type":  {"text/plain"},
	}

	if err := WriteJSONWithHeaders(recorder, http.StatusOK, map[string]string{"ok": "yes"}, headers); err != nil {
		t.Fatalf("WriteJSONWithHeaders() error = %v", err)
	}
	if got := recorder.Header().Get("X-Request-ID"); got != "request-123" {
		t.Fatalf("X-Request-ID = %q", got)
	}
	if got := recorder.Header().Get("Cache-Control"); got != "no-store" {
		t.Fatalf("Cache-Control = %q", got)
	}
	if got := recorder.Header().Get("Content-Type"); got != "application/json; charset=utf-8" {
		t.Fatalf("Content-Type = %q", got)
	}
}

func TestWriteCursorPaginated(t *testing.T) {
	recorder := httptest.NewRecorder()
	if err := WriteCursorPaginated(
		recorder,
		http.StatusOK,
		"Results",
		[]string{"a"},
		"next-token",
		"",
	); err != nil {
		t.Fatalf("WriteCursorPaginated() error = %v", err)
	}

	var got Response
	if err := json.Unmarshal(recorder.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	meta, ok := got.Meta.(map[string]any)
	if !ok || meta["next_cursor"] != "next-token" ||
		meta["has_next"] != true || meta["has_previous"] != false {
		t.Fatalf("cursor metadata = %#v", got.Meta)
	}
}

func TestWritePaginatedWithLinks(t *testing.T) {
	request := httptest.NewRequest(
		http.MethodGet,
		"/users?search=blue&page=2&per_page=5",
		nil,
	)
	recorder := httptest.NewRecorder()

	if err := WritePaginatedWithLinks(
		recorder,
		request,
		http.StatusOK,
		"Users",
		[]string{"Amina"},
		2,
		5,
		12,
	); err != nil {
		t.Fatalf("WritePaginatedWithLinks() error = %v", err)
	}

	link := recorder.Header().Get("Link")
	for _, want := range []string{
		"</users?page=1&per_page=5&search=blue>; rel=\"first\"",
		"</users?page=1&per_page=5&search=blue>; rel=\"prev\"",
		"</users?page=3&per_page=5&search=blue>; rel=\"next\"",
		"</users?page=3&per_page=5&search=blue>; rel=\"last\"",
	} {
		if !strings.Contains(link, want) {
			t.Errorf("Link header %q does not contain %q", link, want)
		}
	}
}

func TestRequestIDMiddleware(t *testing.T) {
	var contextID string
	handler := RequestIDMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contextID = RequestIDFromContext(r.Context())
		w.WriteHeader(http.StatusNoContent)
	}))
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))

	headerID := recorder.Header().Get(RequestIDHeader)
	if len(headerID) != 32 || contextID != headerID {
		t.Fatalf("header ID = %q, context ID = %q", headerID, contextID)
	}
}

func TestWriteValidationError(t *testing.T) {
	recorder := httptest.NewRecorder()
	fields := ValidationErrors{"email": {"is required", "must be valid"}}

	if err := WriteValidationError(recorder, "Invalid input", fields); err != nil {
		t.Fatalf("WriteValidationError() error = %v", err)
	}
	if recorder.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusUnprocessableEntity)
	}

	var got Response
	if err := json.Unmarshal(recorder.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if got.Success || !reflect.DeepEqual(got.Errors, map[string]any{
		"email": []any{"is required", "must be valid"},
	}) {
		t.Fatalf("response = %+v", got)
	}
}

func TestWritePaginated(t *testing.T) {
	recorder := httptest.NewRecorder()
	data := []string{"a", "b"}

	if err := WritePaginated(recorder, http.StatusOK, "Results", data, 2, 2, 5); err != nil {
		t.Fatalf("WritePaginated() error = %v", err)
	}

	var got struct {
		Success bool           `json:"success"`
		Meta    PaginationMeta `json:"meta"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	want := PaginationMeta{
		Page:        2,
		PerPage:     2,
		TotalItems:  5,
		TotalPages:  3,
		HasNext:     true,
		HasPrevious: true,
	}
	if !got.Success || got.Meta != want {
		t.Fatalf("response = %+v, want metadata %+v", got, want)
	}
}

func TestNewPaginationMetaRejectsInvalidInput(t *testing.T) {
	tests := []struct {
		name       string
		page       int
		perPage    int
		totalItems int64
	}{
		{name: "page below one", page: 0, perPage: 10, totalItems: 0},
		{name: "page size below one", page: 1, perPage: 0, totalItems: 0},
		{name: "negative total", page: 1, perPage: 10, totalItems: -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := NewPaginationMeta(tt.page, tt.perPage, tt.totalItems); err == nil {
				t.Fatal("NewPaginationMeta() error = nil, want error")
			}
		})
	}
}
