package responce

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// ProblemDetails represents an RFC 9457 problem details response.
type ProblemDetails struct {
	Type       string
	Title      string
	Status     int
	Detail     string
	Instance   string
	Extensions map[string]any
}

// MarshalJSON encodes the standard problem members alongside extensions.
func (p ProblemDetails) MarshalJSON() ([]byte, error) {
	members := make(map[string]any, 5+len(p.Extensions))
	problemType := p.Type
	if problemType == "" {
		problemType = "about:blank"
	}
	members["type"] = problemType
	if p.Title != "" {
		members["title"] = p.Title
	}
	if p.Status != 0 {
		members["status"] = p.Status
	}
	if p.Detail != "" {
		members["detail"] = p.Detail
	}
	if p.Instance != "" {
		members["instance"] = p.Instance
	}

	for name, value := range p.Extensions {
		switch name {
		case "type", "title", "status", "detail", "instance":
			return nil, fmt.Errorf("problem extension %q conflicts with a standard member", name)
		}
		members[name] = value
	}
	return json.Marshal(members)
}

// WriteProblem writes an RFC 9457 problem response using application/problem+json.
// A zero ProblemDetails.Status is filled from statusCode; a different non-zero
// status is rejected to keep the body and HTTP response consistent.
func WriteProblem(w http.ResponseWriter, statusCode int, problem ProblemDetails) error {
	if problem.Status != 0 && problem.Status != statusCode {
		err := fmt.Errorf("problem status %d does not match HTTP status %d", problem.Status, statusCode)
		logf("responce: %v", err)
		return err
	}
	problem.Status = statusCode
	return writeJSONWithContentType(w, statusCode, problem, nil, "application/problem+json")
}
