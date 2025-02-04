package elastic

import (
	"bytes"
	"context"
	"github.com/goccy/go-json"
	"go-es/internal/errors"
	"net/http"
)

// UpdateIndexMappings updates the mappings of an existing index.
//
// It sends an update mappings request to the Elasticsearch server with the specified index name
// and payload, and returns an error if the request fails.
//
// The payload should be the JSON representation of the mappings to be set.
func (e *elastic) UpdateIndexMappings(ctx context.Context, index string, payload []byte) error {
	// Send update mappings request
	res, err := e.client.Indices.PutMapping(
		[]string{index},
		bytes.NewReader(payload),
		e.client.Indices.PutMapping.WithContext(ctx),
	)
	if err != nil {
		return errors.ElasticsearchError{
			StatusCode: http.StatusInternalServerError,
			Message:    "failed to connect to elastic server",
			Details:    err,
			Type:       errors.ConnectionError,
		}
	}
	defer res.Body.Close()

	// Check if the response is an error
	if res.IsError() {
		return errors.ParseElasticsearchError(res, "failed to update index mappings")
	}

	return nil
}

// UpdateAlias updates the alias to point to the new index.
//
// It sends an update alias request to the Elasticsearch server with the specified alias name
// and the new index name, and returns an error if the request fails.
func (e *elastic) UpdateAlias(ctx context.Context, alias, newIndex string) error {
	// Update alias request body
	aliasBody := map[string]interface{}{
		"actions": []map[string]interface{}{
			{
				"add": map[string]interface{}{
					"index": newIndex,
					"alias": alias,
				},
			},
		},
	}

	// Marshal the request body to JSON
	aliasPayload, err := json.Marshal(aliasBody)
	if err != nil {
		return errors.ElasticsearchError{
			StatusCode: http.StatusInternalServerError,
			Message:    "failed to marshal alias body",
			Details:    err,
			Type:       errors.MarhshalingError,
		}
	}

	// Send update alias request
	res, err := e.client.Indices.UpdateAliases(
		bytes.NewReader(aliasPayload),
		e.client.Indices.UpdateAliases.WithContext(ctx),
	)
	if err != nil {
		return errors.ElasticsearchError{
			StatusCode: http.StatusInternalServerError,
			Message:    "failed to connect to elastic server",
			Details:    err,
			Type:       errors.ConnectionError,
		}
	}
	defer res.Body.Close()

	// Check if the response is an error
	if res.IsError() {
		return errors.ParseElasticsearchError(res, "failed to update alias")
	}

	return nil
}
