package response

// Pagination defines pagination options for search results
type Pagination struct {
	From int `json:"from,omitempty"` // Starting offset
	Size int `json:"size,omitempty"` // Number of documents to return
}

// SearchResponse defines the response structure for search results
type SearchResponse struct {
	Total     int              `json:"total"`     // Total number of matching documents
	MaxScore  float64          `json:"max_score"` // Maximum score of the matching documents
	Documents []SearchDocument `json:"documents"` // List of matching documents
}

// SearchDocument defines the structure of a single document in the search results
type SearchDocument struct {
	ID    string                 `json:"id"`    // Document ID
	Score float64                `json:"score"` // Document score
	Data  map[string]interface{} `json:"data"`  // Document data
}

// SuggestResponse defines the response structure for autocomplete suggestions
type SuggestResponse struct {
	Suggestions []string `json:"suggestions"` // List of suggestions
}

// ListDocumentsResponse defines the response structure for listing documents
type ListDocumentsResponse struct {
	Total     int        `json:"total"`     // Total number of documents
	Documents []Document `json:"documents"` // List of documents
}

// Document defines the structure of a single document in the response
type Document struct {
	ID   string                 `json:"id"`   // Document ID
	Data map[string]interface{} `json:"data"` // Document data
}
