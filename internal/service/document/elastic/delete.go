package elastic

import (
	"context"
	"go-es/internal/errors"
	"net/http"
)

// DeleteDocument deletes a document from an Elasticsearch index.
//
// It will construct a delete request to the Elasticsearch server and return
// an error if the request fails.
func (e *elastic) DeleteDocument(ctx context.Context, index string, docId string) error {
	res, err := e.client.Delete(index, docId, e.client.Delete.WithContext(ctx))
	if err != nil {
		return errors.ElasticsearchError{
			StatusCode: http.StatusInternalServerError,
			Message:    "failed to connect to elastic server",
			Details:    err,
			Type:       errors.ConnectionError,
		}
	}
	defer res.Body.Close()

	if res.IsError() {
		return errors.ParseElasticsearchError(res, "failed to delete document")
	}
	return nil
}
