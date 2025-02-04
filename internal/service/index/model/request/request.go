package request

import "go-es/internal/service/index/model"

// UpdateIndexRequest defines the request structure for updating an index.
type UpdateIndexRequest struct {
	Fields map[string]model.FieldConfig `json:"fields" binding:"required"` // Updated fields configuration
}
