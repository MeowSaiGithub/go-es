package elastic

import (
	"context"
	"github.com/elastic/go-elasticsearch/v8/esutil"
	"github.com/goccy/go-json"
	"go-es/internal/errors"
	"net/http"
)

// ListAllDocuments retrieves all documents from an Elasticsearch index.
//
// The function returns a map containing the search results or an error if the operation fails.
func (e *elastic) ListAllDocuments(ctx context.Context, index string, query map[string]interface{}) (map[string]interface{}, error) {
	// Execute the search request
	res, err := e.client.Search(
		e.client.Search.WithContext(ctx),
		e.client.Search.WithIndex(index),
		e.client.Search.WithBody(esutil.NewJSONReader(query)),
	)
	if err != nil {
		// Return a connection error if the request fails
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
		return nil, errors.ParseElasticsearchError(res, "failed to fetch data")
	}

	// Parse the response body
	var searchResult map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&searchResult); err != nil {
		// Return a decoding error if the response cannot be parsed
		return nil, errors.ElasticsearchError{
			StatusCode: http.StatusInternalServerError,
			Message:    "failed to parse response",
			Details:    err,
			Type:       errors.DecodeError,
		}
	}

	// Return the parsed search results
	return searchResult, nil
}
