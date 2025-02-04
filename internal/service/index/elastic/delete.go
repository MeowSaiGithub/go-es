package elastic

import (
	"context"
	"go-es/internal/errors"
	"net/http"
)

// DeleteIndex deletes a index in Elasticsearch.
//
// It sends a index deletion request to the Elasticsearch server with the specified index name,
// and returns an error if the request fails.
func (e *elastic) DeleteIndex(ctx context.Context, index string) error {
	res, err := e.client.Indices.Delete([]string{index}, e.client.Indices.Delete.WithContext(ctx))
	if err != nil {
		return errors.ElasticsearchError{
			StatusCode: http.StatusInternalServerError,
			Message:    "failed to connect to elastic client",
			Details:    err,
			Type:       errors.ConnectionError,
		}
	}
	defer res.Body.Close()
	if res.IsError() {
		return errors.ParseElasticsearchError(res, "failed to delete elastic client")
	}
	return nil
}
