package handler

import (
	"errors"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/gin-gonic/gin"
	"github.com/goccy/go-json"
	cErr "go-es/internal/errors"
	"go-es/internal/response"
	"go-es/internal/service/document/elastic"
	"go-es/internal/service/document/model/request"
	response2 "go-es/internal/service/document/model/response"
	"go-es/logger"
	"net/http"
	"time"
)

// AutoComplete provides autocomplete suggestions
//
// It takes the alias name as a parameter from the URI and uses it to query
// the Elasticsearch cluster for autocomplete suggestions. It returns a JSON
// response containing the suggestions.
func AutoComplete(e *elasticsearch.Client) func(c *gin.Context) {
	return func(c *gin.Context) {
		logs := logger.GetLogger(c)

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
		var req request.SuggestRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			logs.Error().Err(err).Msg("invalid request payload")
			response.SendErrorResponse(c, response.ErrResponse{
				Code:    http.StatusBadRequest,
				Message: "invalid request payload",
				Details: err.Error(),
				Type:    cErr.BadRequestError.String(),
			})
			return
		}

		// Build suggestion query
		query := buildSuggestQuery(req.Field, req.Input)

		// Convert query to JSON
		payload, err := json.Marshal(query)
		if err != nil {
			logs.Error().Err(err).Msg("failed to marshal auto-complete body")
			response.SendErrorResponse(c, response.ErrResponse{
				Code:    http.StatusBadRequest,
				Message: "failed to marshal auto-complete body",
				Details: err.Error(),
				Type:    cErr.BadRequestError.String(),
			})
			return
		}

		// Call AutoComplete function
		ec := elastic.New(e)
		result, err := ec.AutoComplete(c.Request.Context(), alias, payload)
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
			logs.Error().Err(err).Msg("failed to search auto-complete")
			response.SendErrorResponse(c, response.ErrResponse{
				Code:    http.StatusInternalServerError,
				Message: "failed to search auto-complete",
				Details: err.Error(),
				Type:    cErr.ServerError.String(),
			})
			return
		}

		logs.Info().Msg("suggestions retrieved successfully")
		// Return suggestions
		c.JSON(http.StatusOK, response.SuccessResponse[response2.SuggestResponse]{
			Ts:      time.Now(),
			Code:    http.StatusOK,
			Message: "suggestions retrieved successfully",
			Data:    *result,
		})
	}
}

// buildSuggestQuery constructs the Elasticsearch suggestion query
//
// It takes the field and input as arguments and returns a JSON payload
// that can be used to query the Elasticsearch cluster for autocomplete
// suggestions.
func buildSuggestQuery(field, input string) map[string]interface{} {
	return map[string]interface{}{
		"suggest": map[string]interface{}{
			"text": input, // ✅ Fix: Use "text" instead of "prefix"
			"name_suggestion": map[string]interface{}{
				"completion": map[string]interface{}{
					"field": field + ".suggest", // ✅ Fix: Target the "suggest" completion field
				},
			},
		},
	}
}
