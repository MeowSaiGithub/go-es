package elastic

import (
	"bytes"
	"context"
	"github.com/goccy/go-json"
	"go-es/internal/errors"
	"net/http"
)

// AddBatchData adds a batch of documents to an Elasticsearch index. It will
// construct a bulk request from the input documents and send it to the
// Elasticsearch server.
//
// The documents are expected to be in the form of a slice of maps, where each
// map represents a document to be added. The documents are marshaled to JSON
// and concatenated into a single request body, with each document separated
// by a newline.
func (e *elastic) AddBatchData(ctx context.Context, indexName string, data []map[string]interface{}) error {
	var buf bytes.Buffer

	// Construct bulk request body
	for _, doc := range data {
		// Add the action meta-data for each document
		meta := map[string]interface{}{
			"index": map[string]interface{}{
				"_index": indexName,
			},
		}
		metaJSON, err := json.Marshal(meta)
		if err != nil {
			return errors.ElasticsearchError{
				StatusCode: http.StatusBadRequest,
				Message:    "failed to marshal metadata",
				Details:    err,
				Type:       errors.MarhshalingError,
			}
		}
		buf.Write(metaJSON)
		buf.WriteByte('\n')

		// Add the document itself
		docJSON, err := json.Marshal(doc)
		if err != nil {
			return errors.ElasticsearchError{
				StatusCode: http.StatusBadRequest,
				Message:    "failed to marshal document",
				Details:    err,
				Type:       errors.MarhshalingError,
			}
		}
		buf.Write(docJSON)
		buf.WriteByte('\n')
	}

	// Perform the bulk request
	res, err := e.client.Bulk(bytes.NewReader(buf.Bytes()), e.client.Bulk.WithContext(ctx))
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
		return errors.ParseElasticsearchError(res, "bulk request failed")
	}

	return nil
}
