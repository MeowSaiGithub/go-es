package model

// FieldConfig defines the configuration for a field in the index
type FieldConfig struct {
	Type           string                 `json:"type"`
	Analyzer       string                 `json:"analyzer,omitempty"`
	SearchAnalyzer string                 `json:"search_analyzer,omitempty"`
	Autocomplete   bool                   `json:"autocomplete,omitempty"`
	Search         bool                   `json:"search,omitempty"`
	Properties     map[string]FieldConfig `json:"properties,omitempty"` // For nested fields
}

// Index defines the request structure for creating an index
type Index struct {
	Name   string                 `json:"index"`
	Fields map[string]FieldConfig `json:"fields"`
}
