package responce

import (
	"fmt"
	"math"
)

// PaginationMeta describes the requested page and the full result set.
type PaginationMeta struct {
	Page        int   `json:"page"`
	PerPage     int   `json:"per_page"`
	TotalItems  int64 `json:"total_items"`
	TotalPages  int64 `json:"total_pages"`
	HasNext     bool  `json:"has_next"`
	HasPrevious bool  `json:"has_previous"`
}

// CursorPaginationMeta contains opaque cursors supplied by the caller.
type CursorPaginationMeta struct {
	NextCursor     string `json:"next_cursor,omitempty"`
	PreviousCursor string `json:"previous_cursor,omitempty"`
	HasNext        bool   `json:"has_next"`
	HasPrevious    bool   `json:"has_previous"`
}

// NewPaginationMeta validates pagination inputs and calculates page metadata.
func NewPaginationMeta(page, perPage int, totalItems int64) (PaginationMeta, error) {
	if page < 1 {
		return PaginationMeta{}, fmt.Errorf("page must be at least 1")
	}
	if perPage < 1 {
		return PaginationMeta{}, fmt.Errorf("perPage must be at least 1")
	}
	if totalItems < 0 {
		return PaginationMeta{}, fmt.Errorf("totalItems cannot be negative")
	}

	pageSize := int64(perPage)
	if int(pageSize) != perPage {
		return PaginationMeta{}, fmt.Errorf("perPage is too large")
	}

	totalPages := totalItems / pageSize
	if totalItems%pageSize != 0 {
		totalPages++
	}

	return PaginationMeta{
		Page:        page,
		PerPage:     perPage,
		TotalItems:  totalItems,
		TotalPages:  totalPages,
		HasNext:     int64(page) < totalPages,
		HasPrevious: page > 1,
	}, nil
}

// Offset returns the zero-based SQL-style offset for this page.
func (p PaginationMeta) Offset() (int64, error) {
	if p.Page < 1 || p.PerPage < 1 {
		return 0, fmt.Errorf("page and perPage must be at least 1")
	}

	page := int64(p.Page)
	perPage := int64(p.PerPage)
	if page > 1 && perPage > math.MaxInt64/(page-1) {
		return 0, fmt.Errorf("pagination offset overflows int64")
	}
	return (page - 1) * perPage, nil
}
