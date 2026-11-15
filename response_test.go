package responce

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

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
