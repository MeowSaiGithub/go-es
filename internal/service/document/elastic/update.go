package elastic

import (
	"bytes"
	"context"
	"go-es/internal/errors"
	"net/http"
)

// UpdateDocument updates a document in an Elasticsearch index
//
// It will construct an update request to the Elasticsearch server and return
// an error if the request fails.
//
// The body should be a JSON byte slice that represents the document to be
// updated.
//
// The 'doc' field in the JSON payload is the document to update and the
// 'doc_as_upsert' field is a boolean that is used to specify whether the update
// should be an upsert.
func (e *elastic) UpdateDocument(ctx context.Context, index string, docID string, body []byte) error {
	res, err := e.client.Update(index, docID, bytes.NewReader(body), e.client.Update.WithContext(ctx))
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
		return errors.ParseElasticsearchError(res, "failed to update document")
	}

	return nil
}
