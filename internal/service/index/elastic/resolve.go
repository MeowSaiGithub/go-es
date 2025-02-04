package elastic

import (
	"context"
	"github.com/goccy/go-json"
	"go-es/internal/errors"
	"net/http"
)

// ResolveAlias resolves an alias to the actual index name.
//
// It sends a GetAlias request to the Elasticsearch server using the provided alias name
// and returns the resolved index name or an error if the operation fails.
func (e *elastic) ResolveAlias(ctx context.Context, alias string) (string, error) {
	// Send a GetAlias request to Elasticsearch
	res, err := e.client.Indices.GetAlias(
		e.client.Indices.GetAlias.WithContext(ctx),
		e.client.Indices.GetAlias.WithName(alias),
	)
	if err != nil {
		return "", errors.ElasticsearchError{
			StatusCode: http.StatusInternalServerError,
			Message:    "failed to connect to elastic server",
			Details:    err,
			Type:       errors.ConnectionError,
		}
	}
	defer res.Body.Close()

	// Check if the response contains an error
	if res.IsError() {
		if res.StatusCode == http.StatusNotFound {
			return "", errors.ElasticsearchError{
				StatusCode: http.StatusNotFound,
				Message:    "alias not found",
				Details:    err,
				Type:       errors.NotFoundError,
			}
		}
		return "", errors.ParseElasticsearchError(res, "failed to resolve alias")
	}

	// Decode the response body into a map
	var result map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return "", errors.ElasticsearchError{
			StatusCode: http.StatusInternalServerError,
			Message:    "failed to decode response",
			Details:    err,
			Type:       errors.DecodeError,
		}
	}

	// Extract and return the index name from the alias
	for indexName := range result {
		return indexName, nil
	}

	// Return error if no index name found
	return "", nil
}
