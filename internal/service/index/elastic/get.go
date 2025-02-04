package elastic

import (
	"context"
	"github.com/goccy/go-json"
	"go-es/internal/errors"
	"net/http"
)

// GetIndex retrieves details of a single index from Elasticsearch.
//
// It sends a get index request to the Elasticsearch server with the specified index name,
// and returns the index details in a map[string]interface{} format, or an error if the request fails.
func (e *elastic) GetIndex(ctx context.Context, index string) (map[string]interface{}, error) {
	res, err := e.client.Indices.Get([]string{index}, e.client.Indices.Get.WithContext(ctx))
	if err != nil {
		return nil, errors.ElasticsearchError{
			StatusCode: http.StatusInternalServerError,
			Message:    "failed to connect to elastic client",
			Details:    err,
			Type:       errors.ConnectionError,
		}
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, errors.ParseElasticsearchError(res, "failed to get index info")
	}

	// Decode the response
	var info map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&info); err != nil {
		return nil, errors.ElasticsearchError{
			StatusCode: http.StatusInternalServerError,
			Message:    "failed to decode response",
			Details:    err,
			Type:       errors.DecodeError,
		}
	}

	return info, nil
}
