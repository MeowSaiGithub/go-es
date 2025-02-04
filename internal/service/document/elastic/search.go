package elastic

import (
	"bytes"
	"context"
	"github.com/goccy/go-json"
	"go-es/internal/errors"
	"go-es/internal/service/document/model/response"
	"net/http"
)

// Search performs a search query on an Elasticsearch index and returns the search results.
//
// The function returns a SearchResponse containing the total hits, max score, and a list of documents,
// or an error if the search operation fails.
func (e *elastic) Search(ctx context.Context, indexName string, payload []byte) (*response.SearchResponse, error) {
	// Execute the search query against the Elasticsearch client
	res, err := e.client.Search(
		e.client.Search.WithContext(ctx),
		e.client.Search.WithIndex(indexName),
		e.client.Search.WithBody(bytes.NewReader(payload)),
	)
	if err != nil {
		return nil, errors.ElasticsearchError{
			StatusCode: http.StatusInternalServerError,
			Message:    "failed to connect to elastic server",
			Details:    err,
			Type:       errors.ConnectionError,
		}
	}
	defer res.Body.Close()

	// Check if the search response contains an error
	if res.IsError() {
		return nil, errors.ParseElasticsearchError(res, "failed to search")
	}

	// Define a structure to parse the search response
	var esResponse struct {
		Hits struct {
			Total struct {
				Value int `json:"value"`
			} `json:"total"`
			MaxScore float64 `json:"max_score"`
			Hits     []struct {
				ID     string                 `json:"_id"`
				Score  float64                `json:"_score"`
				Source map[string]interface{} `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	// Decode the search response JSON into the esResponse structure
	if err = json.NewDecoder(res.Body).Decode(&esResponse); err != nil {
		return nil, errors.ElasticsearchError{
			StatusCode: http.StatusInternalServerError,
			Message:    "failed to decode search response",
			Details:    err,
			Type:       errors.DecodeError,
		}
	}

	// Transform esResponse into SearchResponse
	searchResponse := &response.SearchResponse{
		Total:     esResponse.Hits.Total.Value,
		MaxScore:  esResponse.Hits.MaxScore,
		Documents: make([]response.SearchDocument, 0, len(esResponse.Hits.Hits)),
	}

	// Iterate over the hits and append each document to the search response
	for _, hit := range esResponse.Hits.Hits {
		searchResponse.Documents = append(searchResponse.Documents, response.SearchDocument{
			ID:    hit.ID,
			Score: hit.Score,
			Data:  hit.Source,
		})
	}

	return searchResponse, nil
}
