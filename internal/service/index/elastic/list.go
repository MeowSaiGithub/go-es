package elastic

import (
	"context"
	"github.com/goccy/go-json"
	"go-es/internal/errors"
	"net/http"
	"strings"
)

// ListIndices returns a map of indices names to their corresponding aliases.
//
// It calls the Elasticsearch _cat/indices API to get a list of all indices,
// and the _cat/aliases API to get a list of all aliases. It then filters out
// indices starting with '.' (system/internal indices), and returns a map of
// index names to their corresponding aliases.
func (e *elastic) ListIndices(ctx context.Context) (map[string]string, error) {

	// Get all indices (with format="json" to return JSON instead of plain text)
	res, err := e.client.Cat.Indices(e.client.Cat.Indices.WithContext(ctx), e.client.Cat.Indices.WithFormat("json"))
	if err != nil {
		return nil, errors.ElasticsearchError{
			StatusCode: http.StatusInternalServerError,
			Message:    "failed to connect to elastic client",
			Details:    err,
			Type:       errors.ConnectionError,
		}
	}
	defer res.Body.Close()

	// Check if the response is an error
	if res.IsError() {
		return nil, errors.ParseElasticsearchError(res, "failed to get indices")
	}

	// Parse the response body into a JSON array
	var rawIndices []map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&rawIndices); err != nil {
		return nil, errors.ElasticsearchError{
			StatusCode: http.StatusInternalServerError,
			Message:    "failed to parse response",
			Details:    err,
			Type:       errors.DecodeError,
		}
	}

	aliasRes, err := e.client.Cat.Aliases(e.client.Cat.Aliases.WithContext(ctx), e.client.Cat.Aliases.WithFormat("json"))
	if err != nil {
		return nil, errors.ElasticsearchError{
			StatusCode: http.StatusInternalServerError,
			Message:    "failed to connect to elastic client",
			Details:    err,
			Type:       errors.ConnectionError,
		}
	}
	defer aliasRes.Body.Close()

	// Check if the response is an error
	if aliasRes.IsError() {
		return nil, errors.ParseElasticsearchError(res, "failed to get aliases")
	}

	var rawAliases []map[string]interface{}
	if err := json.NewDecoder(aliasRes.Body).Decode(&rawAliases); err != nil {
		return nil, errors.ElasticsearchError{
			StatusCode: http.StatusInternalServerError,
			Message:    "failed to parse alias response",
			Details:    err,
			Type:       errors.DecodeError,
		}
	}

	// Filter out indices starting with '.'
	indiceAliasMap := make(map[string]string)
	for _, index := range rawIndices {
		// Extract the index name
		indexName := index["index"].(string)

		// Exclude indices starting with '.' (system/internal indices)
		if !strings.HasPrefix(indexName, ".") {
			indiceAliasMap[indexName] = ""
		}
	}

	for _, a := range rawAliases {
		indexName := a["index"].(string)
		aliasName := a["alias"].(string)
		if !strings.HasPrefix(indexName, ".") {
			indiceAliasMap[indexName] = aliasName
		}

	}

	return indiceAliasMap, nil
}
