package errors

import (
	"fmt"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/goccy/go-json"
)

type ErrType string

func (e ErrType) String() string {
	return string(e)
}

// ErrType represents a string alias for different types of errors.
const (
	// ConnectionError indicates an error due to connection issues.
	ConnectionError ErrType = "connection_error"
	// ParseError indicates an error during parsing.
	ParseError ErrType = "parse_error"
	// NoResponseError indicates no response from the server.
	NoResponseError ErrType = "no_response"
	// DecodeError indicates an error during decoding.
	DecodeError ErrType = "decode_error"
	// ResourceAlreadyExistsException indicates a resource already exists.
	ResourceAlreadyExistsException ErrType = "resource_already_exists_exception"
	// IndexNotFoundExc indicates that the index was not found.
	IndexNotFoundExc ErrType = "index_not_found_exception"
	// IllegalArgumentException indicates an invalid argument was provided.
	IllegalArgumentException ErrType = "illegal_argument_exception"
	// ValidationException indicates a validation error.
	ValidationException ErrType = "validation_exception"
	// MarhshalingError indicates an error during marshaling.
	MarhshalingError ErrType = "marshaling_error"
	// BadRequestError indicates a bad request.
	BadRequestError ErrType = "bad_request"
	// ServerError indicates an internal server error.
	ServerError ErrType = "server_error"
	// NotFoundError indicates a resource was not found.
	NotFoundError ErrType = "not_found"
	// UnauthorizedError indicates an unauthorized request.
	UnauthorizedError ErrType = "unauthorized"
)

// ElasticsearchError represents a structured error object returned by the
// Elasticsearch client.
type ElasticsearchError struct {
	StatusCode int     `json:"status_code"` // HTTP status code
	Message    string  `json:"message"`     // Simplified user-friendly message
	Details    error   `json:"details"`     // Detailed technical message for debugging
	Type       ErrType `json:"type"`        // Elasticsearch error type
}

// Error implements the error interface for ElasticsearchError.
func (e ElasticsearchError) Error() string {
	return fmt.Sprintf("Status: %d, Type: %s, Message: %s, Details: %s", e.StatusCode, e.Type, e.Message, e.Details)
}

// Is allows comparison of errors using errors.Is().
func (e ElasticsearchError) Is(target error) bool {
	t, ok := target.(ElasticsearchError)
	if !ok {
		return false
	}
	// Compare the error types to see if they match.
	return e.Type == t.Type
}

// As allows assertion of errors using errors.As().
func (e ElasticsearchError) As(target any) bool {
	t, ok := target.(*ElasticsearchError)
	if !ok {
		return false
	}
	*t = e
	return true
}

// ParseElasticsearchError parses the error response from Elasticsearch into a structured error object.
// It tries to extract the error type and a user-friendly message from the response.
// If parsing fails, it falls back to a default error message.
func ParseElasticsearchError(res *esapi.Response, defaultMessage string) error {
	if res == nil {
		return ElasticsearchError{
			StatusCode: 0,
			Message:    "No response from Elasticsearch",
			Details:    fmt.Errorf(defaultMessage),
			Type:       NoResponseError,
		}
	}

	defer res.Body.Close()

	var errorResponse map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&errorResponse); err != nil {
		return ElasticsearchError{
			StatusCode: res.StatusCode,
			Message:    defaultMessage,
			Details:    fmt.Errorf("failed to decode error response"),
			Type:       DecodeError,
		}
	}

	errType := ""
	errReason := defaultMessage
	if errMap, ok := errorResponse["error"].(map[string]interface{}); ok {
		if errTypeField, found := errMap["type"].(string); found {
			errType = errTypeField
		}
		if reasonField, found := errMap["reason"].(string); found {
			errReason = reasonField
		}
	}

	var userMessage string
	switch ErrType(errType) {
	case ResourceAlreadyExistsException:
		userMessage = "The specified index already exists."
	case IndexNotFoundExc:
		userMessage = "The requested index does not exist."
	case IllegalArgumentException:
		userMessage = "Invalid request. Please check your input."
	case ValidationException:
		userMessage = "Validation error. Please verify your request payload."
	default:
		userMessage = defaultMessage
	}

	return ElasticsearchError{
		StatusCode: res.StatusCode,
		Message:    userMessage,
		Details:    fmt.Errorf(errReason),
		Type:       ErrType(errType),
	}
}
