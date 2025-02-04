package elastic

import (
	"context"
	"fmt"
	"github.com/goccy/go-json"
	logs "github.com/rs/zerolog/log"
	"go-es/internal/errors"
	"net/http"
	"strings"
	"time"
)

const (
	MaxScrollIterations = 100
	MaxScrollTime       = 2 * time.Minute
)

// ExportDocuments exports documents from an Elasticsearch index
//
// This function uses the search API with scrolling to efficiently fetch documents
// from Elasticsearch. It will scroll until there are no more documents left or
// the maximum scroll iterations is reached (100 by default).
//
// The function returns a slice of documents, each of which is represented as a map
// of strings to arbitrary values. The documents are ordered in the same order
// as they were returned by the search query.
func (e *elastic) ExportDocuments(ctx context.Context, index string, query map[string]interface{}) ([]map[string]interface{}, error) {
	var documents []map[string]interface{}

	// Build the query
	queryBody := map[string]interface{}{
		"query": query,
	}
	if query == nil {
		queryBody["query"] = map[string]interface{}{
			"match_all": map[string]interface{}{},
		}
	}

	queryBytes, err := json.Marshal(queryBody)
	if err != nil {
		return nil, errors.ElasticsearchError{
			StatusCode: http.StatusInternalServerError,
			Message:    "failed to serialize query",
			Details:    err,
			Type:       errors.MarhshalingError,
		}
	}

	// Initial search request
	res, err := e.client.Search(
		e.client.Search.WithContext(ctx),
		e.client.Search.WithIndex(index),
		e.client.Search.WithBody(strings.NewReader(string(queryBytes))),
		e.client.Search.WithScroll(MaxScrollTime), // Add scroll context for larger datasets
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

	if res.IsError() {
		return nil, errors.ParseElasticsearchError(res, "failed to search documents for export")
	}

	var searchResult map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&searchResult); err != nil {
		return nil, errors.ElasticsearchError{
			StatusCode: http.StatusInternalServerError,
			Message:    "failed to decode elastic response",
			Details:    err,
			Type:       errors.DecodeError,
		}
	}

	// Get total hits
	totalHits, ok := searchResult["hits"].(map[string]interface{})["total"].(map[string]interface{})["value"].(float64)
	if !ok {
		return nil, errors.ElasticsearchError{
			StatusCode: http.StatusInternalServerError,
			Message:    "failed to determine hits from response",
			Details:    fmt.Errorf("failed to determine hits from response"),
			Type:       errors.ParseError,
		}
	}

	// Process initial batch of documents
	for _, hit := range searchResult["hits"].(map[string]interface{})["hits"].([]interface{}) {
		source := hit.(map[string]interface{})["_source"]
		documents = append(documents, source.(map[string]interface{}))
	}

	// If total hits fit within the first batch, no need for scroll
	if len(documents) >= int(totalHits) {
		return documents, nil
	}

	// Otherwise, proceed with scrolling
	scrollID, ok := searchResult["_scroll_id"].(string)
	if !ok {
		return nil, errors.ElasticsearchError{
			StatusCode: http.StatusInternalServerError,
			Message:    "failed to get scroll_id from response",
			Details:    fmt.Errorf("failed to get scroll_id from response"),
			Type:       errors.ParseError,
		}
	}

	for i := 0; i < MaxScrollIterations; i++ {
		res, err := e.client.Scroll(
			e.client.Scroll.WithContext(ctx),
			e.client.Scroll.WithScrollID(scrollID),
			e.client.Scroll.WithScroll(MaxScrollTime),
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

		if res.IsError() {
			return nil, errors.ParseElasticsearchError(res, "failed to scroll documents to export")
		}

		if err := json.NewDecoder(res.Body).Decode(&searchResult); err != nil {
			return nil, errors.ElasticsearchError{
				StatusCode: http.StatusInternalServerError,
				Message:    "failed to decode elastic response",
				Details:    err,
				Type:       errors.DecodeError,
			}
		}

		// Process batch
		hits := searchResult["hits"].(map[string]interface{})["hits"].([]interface{})
		if len(hits) == 0 {
			break // No more results, end scrolling
		}

		for _, hit := range hits {
			source := hit.(map[string]interface{})["_source"]
			documents = append(documents, source.(map[string]interface{}))
		}

		// Update scroll ID for the next batch
		scrollID, ok = searchResult["_scroll_id"].(string)
		if !ok || scrollID == "" {
			break // Missing or empty scroll_id, end scrolling
		}
	}

	// Clear scroll context
	_, err = e.client.ClearScroll(
		e.client.ClearScroll.WithScrollID(scrollID),
		e.client.ClearScroll.WithContext(ctx),
	)
	if err != nil {
		logs.Warn().Err(err).Msg("failed to clear scroll")
	}

	return documents, nil
}
