package elastic

import (
	"context"
	"github.com/goccy/go-json"
	"go-es/internal/errors"
	"net/http"
)

// GetDocumentByID retrieves a document by its ID
//
// It will construct a get request to the Elasticsearch server and return
// an error if the request fails.
func (e *elastic) GetDocumentByID(ctx context.Context, index string, docID string) (map[string]interface{}, error) {
	// Get document
	res, err := e.client.Get(index, docID, e.client.Get.WithContext(ctx))
	if err != nil {
		return nil, errors.ElasticsearchError{
			StatusCode: http.StatusInternalServerError,
			Message:    "failed to connect to elastic server",
			Details:    err,
			Type:       errors.ConnectionError,
		}
	}
	defer res.Body.Close()
	if res.IsError() {
		//return nil, fmt.Errorf("failed to fetch document: %s", res.String())
		return nil, errors.ParseElasticsearchError(res, "failed to fetch document")
	}

	// Parse response
	var doc map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&doc); err != nil {
		return nil, errors.ElasticsearchError{
			StatusCode: http.StatusInternalServerError,
			Message:    "failed to decode response",
			Details:    err,
			Type:       errors.DecodeError,
		}
	}

	return doc, nil
}
