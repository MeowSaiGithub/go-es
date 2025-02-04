package elastic

import (
	"context"
	"go-es/internal/errors"
	"net/http"
)

// CheckIndex checks if an index exists on Elasticsearch.
//
// It sends an exists request to the Elasticsearch server using the provided index name,
// and returns true if the index exists, false if it doesn't exist, or an error if the operation fails.
func (e *elastic) CheckIndex(ctx context.Context, index string) (bool, error) {
	res, err := e.client.Indices.Exists([]string{index}, e.client.Indices.Exists.WithContext(ctx))
	if err != nil {
		return false, errors.ElasticsearchError{
			StatusCode: http.StatusInternalServerError,
			Message:    "failed to connect to elastic server",
			Details:    err,
			Type:       errors.ConnectionError,
		}
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusOK {
		// Index exists
		return true, nil
	}
	// Index does not exist
	return false, errors.ParseElasticsearchError(res, "failed to check for index or does not exist")
}
