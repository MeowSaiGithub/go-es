package elastic

import (
	"bytes"
	"context"
	"go-es/internal/errors"
	"net/http"
)

// CreateIndex creates a new index in Elasticsearch with the specified name and settings.
//
// It sends an index creation request to the Elasticsearch server using the provided index name
// and payload, and returns an error if the operation fails.
func (e *elastic) CreateIndex(ctx context.Context, index string, payload []byte) error {
	// Send index creation request
	res, err := e.client.Indices.Create(
		index,
		e.client.Indices.Create.WithContext(ctx),
		e.client.Indices.Create.WithBody(bytes.NewReader(payload)),
	)
	if err != nil {
		return errors.ElasticsearchError{
			StatusCode: http.StatusInternalServerError,
			Message:    "failed to connect to elastic client",
			Details:    err,
			Type:       errors.ConnectionError,
		}
	}
	defer res.Body.Close()

	// Check if the response contains an error
	if res.IsError() {
		return errors.ParseElasticsearchError(res, "failed to create index")
	}

	return nil
}
