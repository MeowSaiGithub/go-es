package elastic

import (
	"context"
	"github.com/goccy/go-json"
	"go-es/internal/errors"
	"net/http"
	"strings"
)

// ImportDocuments imports documents into Elasticsearch.
//
// This function takes a slice of map[string]interface{} representing the
// documents to be imported. Each map should contain the fields of the
// document.
//
// The documents are bulk imported using the Elasticsearch bulk API. The
// documents are marshaled to JSON and the bulk request body is constructed.
// The request is then sent to the Elasticsearch server.
//
// If the request is successful, the function returns nil. Otherwise, an
// error is returned.
func (e *elastic) ImportDocuments(ctx context.Context, index string, documents []map[string]interface{}) error {
	var bulkData strings.Builder

	for _, doc := range documents {
		meta := map[string]interface{}{
			"index": map[string]interface{}{
				"_index": index,
			},
		}
		metaJSON, _ := json.Marshal(meta)
		docJSON, _ := json.Marshal(doc)

		bulkData.WriteString(string(metaJSON) + "\n")
		bulkData.WriteString(string(docJSON) + "\n")
	}

	res, err := e.client.Bulk(strings.NewReader(bulkData.String()), e.client.Bulk.WithContext(ctx))
	if err != nil {
		return errors.ElasticsearchError{
			StatusCode: http.StatusInternalServerError,
			Message:    "failed to connect to elastic server",
			Details:    err,
			Type:       "connection_error",
		}
	}
	defer res.Body.Close()

	if res.IsError() {
		return errors.ParseElasticsearchError(res, "failed to bulk import")
	}
	return nil
}
