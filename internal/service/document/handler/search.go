package handler

import (
	"errors"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/gin-gonic/gin"
	"github.com/goccy/go-json"
	"github.com/rs/zerolog/log"
	cErr "go-es/internal/errors"
	"go-es/internal/response"
	"go-es/internal/service/document/elastic"
	"go-es/internal/service/document/model/request"
	docResp "go-es/internal/service/document/model/response"
	"net/http"
	"time"
)

// Search searches documents in an Elasticsearch index
//
// This function takes a request body containing a query and optional pagination parameters,
// and returns a JSON response containing the search results.
func Search(e *elasticsearch.Client) func(c *gin.Context) {
	return func(c *gin.Context) {
		logs := log.With().Str("handler", "Search").Logger()

		// Extract alias name from the URI
		alias := c.Param("alias")
		if alias == "" {
			logs.Error().Msg("alias name is required in URI")
			response.SendErrorResponse(c, response.ErrResponse{
				Code:    http.StatusBadRequest,
				Message: "alias name is required in URI",
				Details: "alias name is required in URI",
				Type:    cErr.BadRequestError.String(),
			})
			return
		}

		// Parse the JSON body
		var req request.SearchRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			logs.Error().Err(err).Msg("invalid request payload")
			response.SendErrorResponse(c, response.ErrResponse{
				Code:    http.StatusBadRequest,
				Message: "invalid JSON payload",
				Details: err.Error(),
				Type:    cErr.BadRequestError.String(),
			})
			return
		}

		// Build the query
		query := buildQuery(req)
		if query == nil {
			logs.Error().Msg("no valid query provided")
			response.SendErrorResponse(c, response.ErrResponse{
				Code:    http.StatusBadRequest,
				Message: "no valid query provided",
				Details: "no valid query provided",
				Type:    cErr.BadRequestError.String(),
			})
			return
		}

		// Add pagination
		from := req.Pagination.From
		size := req.Pagination.Size
		if size == 0 {
			size = 10 // Default size
		}

		// Build the search body
		searchBody := map[string]interface{}{
			"query": query,
			"from":  from,
			"size":  size,
		}

		// Add minimum score if specified
		if req.MinScore > 0 {
			searchBody["min_score"] = req.MinScore
		}

		// Perform the search
		bodyBytes, err := json.Marshal(searchBody)
		if err != nil {
			logs.Error().Err(err).Msg("failed to marshal search body")
			response.SendErrorResponse(c, response.ErrResponse{
				Code:    http.StatusBadRequest,
				Message: "failed to marshal search body",
				Details: err.Error(),
				Type:    cErr.BadRequestError.String(),
			})
			return
		}

		ec := elastic.New(e)
		result, err := ec.Search(c.Request.Context(), alias, bodyBytes)
		if err != nil {
			var esErr cErr.ElasticsearchError
			if errors.As(err, &esErr) {
				logs.Error().Err(esErr.Details).Msg(esErr.Message)
				response.SendErrorResponse(c, response.ErrResponse{
					Code:    esErr.StatusCode,
					Message: esErr.Message,
					Details: esErr.Details.Error(),
					Type:    esErr.Type.String(),
				})
				return
			}
			logs.Error().Err(err).Msg("failed to search")
			response.SendErrorResponse(c, response.ErrResponse{
				Code:    http.StatusInternalServerError,
				Message: "failed to search",
				Details: err.Error(),
				Type:    cErr.ServerError.String(),
			})
			return
		}

		logs.Info().Msg("search successful")
		c.JSON(http.StatusOK, response.SuccessResponse[docResp.SearchResponse]{
			Ts:      time.Now(),
			Code:    http.StatusOK,
			Message: "search successful",
			Data:    *result,
		})

	}
}

// buildQuery constructs the Elasticsearch query based on the search request
func buildQuery(req request.SearchRequest) map[string]interface{} {
	var query map[string]interface{}

	if req.MatchAll {
		query = map[string]interface{}{
			"match_all": map[string]interface{}{},
		}
	} else if req.Query != "" {
		// Determine the fields to search in
		searchFields := req.SearchFields
		if len(searchFields) == 0 {
			// Default to searching in all text fields if no fields are specified
			searchFields = []string{"name.fulltext", "description.fulltext"}
		}

		// Use a bool query to combine multiple query types
		query = map[string]interface{}{
			"bool": map[string]interface{}{
				"should": []map[string]interface{}{
					{
						"multi_match": map[string]interface{}{
							"query":  req.Query,
							"fields": searchFields,
							"type":   "phrase", // Prioritize exact phrase matches
							"boost":  2,        // Boost exact phrase matches
						},
					},
					{
						"multi_match": map[string]interface{}{
							"query":     req.Query,
							"fields":    searchFields,
							"fuzziness": "AUTO", // Enable fuzzy matching
							"boost":     1,      // Lower boost for fuzzy matches
						},
					},
				},
				"minimum_should_match": 1, // At least one "should" clause must match
				"filter":               buildFilters(req.Filters),
			},
		}
	} else if len(req.Filters) > 0 {
		query = map[string]interface{}{
			"bool": map[string]interface{}{
				"filter": buildFilters(req.Filters),
			},
		}
	}

	return query
}

// buildFilters constructs the filter clauses for the Elasticsearch query
func buildFilters(filters map[string]interface{}) []map[string]interface{} {
	var filterList []map[string]interface{}
	for key, value := range filters {
		filterList = append(filterList, map[string]interface{}{
			"term": map[string]interface{}{
				key: value,
			},
		})
	}
	return filterList
}
