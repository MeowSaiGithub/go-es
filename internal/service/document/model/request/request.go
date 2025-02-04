package request

import (
	"go-es/internal/service/document/model/response"
	"mime/multipart"
)

// UpdateDocumentRequest defines the request structure for updating a document
type UpdateDocumentRequest struct {
	Data map[string]interface{} `json:"data" binding:"required"` // Fields to update
}

// ExportDocumentsRequest defines the request structure for exporting documents
type ExportDocumentsRequest struct {
	Query map[string]interface{} `json:"query"` // Elasticsearch query for filtering
}

// ImportRequest defines the request structure for importing documents.
type ImportRequest struct {
	Index string                `form:"index" binding:"required"` // Elasticsearch index name
	File  *multipart.FileHeader `form:"file"`                     // JSON file (optional)
	JSON  string                `form:"json"`                     // JSON content (optional)
	Bulk  bool                  `form:"bulk"`                     // Whether the input is in bulk format
}

// AddDataRequest defines the request structure for adding data to an index
type AddDataRequest struct {
	Data []map[string]interface{} `json:"data"` // Array of documents to be added
}

// SearchRequest defines the request structure for searching documents in an index
type SearchRequest struct {
	Query        string                 `json:"query,omitempty"`         // Full-text search query
	Filters      map[string]interface{} `json:"filters,omitempty"`       // Filters by field
	MatchAll     bool                   `json:"match_all,omitempty"`     // Match all documents
	Pagination   response.Pagination    `json:"pagination,omitempty"`    // Pagination options
	MinScore     float64                `json:"min_score,omitempty"`     // Minimum relevance score
	SearchFields []string               `json:"search_fields,omitempty"` // Fields to search in
}

// SuggestRequest defines the request structure for autocomplete suggestions
type SuggestRequest struct {
	Field string `json:"field" binding:"required"` // Field to suggest on
	Input string `json:"input" binding:"required"` // Input text for suggestions
}
