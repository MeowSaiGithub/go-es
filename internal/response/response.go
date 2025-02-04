package response

import (
	"github.com/gin-gonic/gin"
	"time"
)

var d bool

// ErrResponse represents the structure for error responses
type ErrResponse struct {
	Ts      time.Time `json:"ts"`                // Timestamp of the error response
	Code    int       `json:"code"`              // HTTP status code
	Message string    `json:"message"`           // Error message
	Details string    `json:"details,omitempty"` // Additional details about the error
	Type    string    `json:"type"`              // Type of the error
}

// SuccessResponse represents the structure for successful responses
type SuccessResponse[T any] struct {
	Ts      time.Time `json:"ts"`             // Timestamp of the response
	Code    int       `json:"code"`           // HTTP status code
	Message string    `json:"message"`        // Success message
	Data    T         `json:"data,omitempty"` // Data returned in case of success
}

// Init initializes the package with the given details flag.
//
// The details flag determines whether to include additional details in error responses.
func Init(details bool) {
	d = details
}

// SendErrorResponse sends a JSON response with the given error response.
//
// If the details flag was set to false when the package was initialized,
// the details field of the error response is cleared before sending.
func SendErrorResponse(c *gin.Context, resp ErrResponse) {
	resp.Ts = time.Now()
	if !d {
		resp.Details = ""
	}
	c.JSON(resp.Code, resp)
}
