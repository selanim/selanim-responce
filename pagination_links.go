package responce

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// WritePaginatedWithLinks writes a page response and RFC 8288 Link headers.
// It preserves other query parameters and uses "page" and "per_page" for pagination.
func WritePaginatedWithLinks(
	w http.ResponseWriter,
	r *http.Request,
	statusCode int,
	message string,
	data any,
	page int,
	perPage int,
	totalItems int64,
) error {
	if r == nil || r.URL == nil {
		err := fmt.Errorf("request URL is required for pagination links")
		logf("responce: %v", err)
		return err
	}

	meta, err := NewPaginationMeta(page, perPage, totalItems)
	if err != nil {
		logf("responce: create paginated response: %v", err)
		return err
	}

	links := make([]string, 0, 4)
	if meta.TotalPages > 0 {
		if page > 1 {
			links = append(links, pageLink(r.URL, 1, perPage, "first"))
			links = append(links, pageLink(r.URL, int64(page)-1, perPage, "prev"))
		}
		if meta.HasNext {
			links = append(links, pageLink(r.URL, int64(page)+1, perPage, "next"))
		}
		if int64(page) != meta.TotalPages {
			links = append(links, pageLink(r.URL, meta.TotalPages, perPage, "last"))
		}
	}

	headers := make(http.Header)
	if len(links) > 0 {
		headers.Set("Link", strings.Join(links, ", "))
	}

	return WriteJSONWithHeaders(w, statusCode, Response{
		Success: true,
		Message: message,
		Data:    data,
		Meta:    meta,
	}, headers)
}

func pageLink(requestURL *url.URL, page int64, perPage int, relation string) string {
	linkURL := *requestURL
	query := linkURL.Query()
	query.Set("page", strconv.FormatInt(page, 10))
	query.Set("per_page", strconv.Itoa(perPage))
	linkURL.RawQuery = query.Encode()
	return fmt.Sprintf("<%s>; rel=\"%s\"", linkURL.String(), relation)
}
