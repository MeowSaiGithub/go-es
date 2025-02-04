package elastic

import (
	"bytes"
	"context"
	"github.com/goccy/go-json"
	"go-es/internal/errors"
	"net/http"
)

// Reindex copies data from the old index to a new index with updated mappings.
//
// It sends a reindexing request to the Elasticsearch server with the specified old
// index name, new index name, and payload, and returns an error if the operation
// fails.
//
// The payload should be the JSON representation of the mappings to be set for the
// new index.
func (e *elastic) Reindex(ctx context.Context, oldIndex, newIndex string, payload []byte) error {
	// Create the new index
	if err := e.CreateIndex(ctx, newIndex, payload); err != nil {
		return errors.ElasticsearchError{
			StatusCode: http.StatusInternalServerError,
			Message:    "failed to connect to elastic server",
			Details:    err,
			Type:       errors.ConnectionError,
		}
	}

	// Perform reindexing
	reindexBody := map[string]interface{}{
		"source": map[string]interface{}{
			"index": oldIndex,
		},
		"dest": map[string]interface{}{
			"index": newIndex,
		},
	}

	reindexPayload, err := json.Marshal(reindexBody)
	if err != nil {
		return errors.ElasticsearchError{
			StatusCode: http.StatusInternalServerError,
			Message:    "failed to marshal reindex body",
			Details:    err,
			Type:       errors.MarhshalingError,
		}
	}

	res, err := e.client.Reindex(
		bytes.NewReader(reindexPayload),
		e.client.Reindex.WithContext(ctx),
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

	if res.IsError() {
		return errors.ParseElasticsearchError(res, "failed to reindex data")
	}

	return nil
}
