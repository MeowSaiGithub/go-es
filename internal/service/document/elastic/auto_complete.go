package elastic

import (
	"bytes"
	"context"
	"github.com/goccy/go-json"
	"go-es/internal/errors"
	"go-es/internal/service/document/model/response"
	"net/http"
)

// AutoComplete performs an autocomplete suggestion query on an Elasticsearch index.
//
// It sends a search request to the Elasticsearch server using the provided index name and query payload,
// and returns the parsed suggestions or an error if the operation fails.
func (e *elastic) AutoComplete(ctx context.Context, indexName string, payload []byte) (*response.SuggestResponse, error) {
	// Execute the search request
	res, err := e.client.Search(
		e.client.Search.WithContext(ctx),
		e.client.Search.WithIndex(indexName),
		e.client.Search.WithBody(bytes.NewReader(payload)),
	)
	if err != nil {
		return nil, errors.ElasticsearchError{
			StatusCode: http.StatusInternalServerError,
			Message:    "failed to connect to elastic server",
			Details:    err,
			Type:       errors.ConnectionError,
		}
	}
	defer res.Body.Close()

	// Check if the response contains an error
	if res.IsError() {
		return nil, errors.ParseElasticsearchError(res, "failed to perform suggestion query")
	}

	// Parse the suggestion response
	var parsedRes map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&parsedRes); err != nil {
		return nil, errors.ElasticsearchError{
			StatusCode: http.StatusInternalServerError,
			Message:    "failed to parse suggestion response",
			Details:    err,
			Type:       errors.DecodeError,
		}
	}

	// Extract suggestions from the parsed response
	suggestions := []string{}
	if suggest, found := parsedRes["suggest"]; found {
		if nameSuggestion, ok := suggest.(map[string]interface{})["name_suggestion"].([]interface{}); ok {
			for _, s := range nameSuggestion {
				if options, ok := s.(map[string]interface{})["options"].([]interface{}); ok {
					for _, option := range options {
						if text, ok := option.(map[string]interface{})["text"].(string); ok {
							suggestions = append(suggestions, text)
						}
					}
				}
			}
		}
	}

	// Return the suggestion response
	return &response.SuggestResponse{
		Suggestions: suggestions,
	}, nil
}
