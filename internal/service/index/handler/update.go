package handler

import (
	"errors"
	"fmt"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/gin-gonic/gin"
	"github.com/goccy/go-json"
	cErr "go-es/internal/errors"
	"go-es/internal/response"
	"go-es/internal/service/index/elastic"
	"go-es/internal/service/index/model/request"
	"go-es/logger"
	"net/http"
	"time"
)

// UpdateIndex handles updating an existing index or reindexing if necessary.
//
// The handler takes a gin Context, extracts the index alias and the new
// index properties from the request body, and uses the elastic package to
// update the index properties or reindex the data if necessary. It sends an
// appropriate JSON response based on the result of the operation.
func UpdateIndex(e *elasticsearch.Client) func(c *gin.Context) {
	return func(c *gin.Context) {
		logs := logger.GetLogger(c)

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

		var req request.UpdateIndexRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			logs.Error().Err(err).Msg("Invalid request payload")
			response.SendErrorResponse(c, response.ErrResponse{
				Code:    http.StatusBadRequest,
				Message: "Invalid request payload",
				Details: err.Error(),
				Type:    cErr.BadRequestError.String(),
			})
			return
		}

		// Generate properties for the updated alias
		properties, err := generateProperties(req.Fields)
		if err != nil {
			logs.Error().Err(err).Msg("failed to generate fields")
			response.SendErrorResponse(c, response.ErrResponse{
				Code:    http.StatusBadRequest,
				Message: "failed to generate fields",
				Details: err.Error(),
				Type:    cErr.BadRequestError.String(),
			})
			return
		}

		// Build Elasticsearch settings and mappings
		indexBody := map[string]interface{}{
			"settings": map[string]interface{}{
				"analysis": map[string]interface{}{
					"analyzer": map[string]interface{}{
						"autocomplete_analyzer": map[string]interface{}{
							"tokenizer": "edge_ngram",
							"filter": []string{
								"lowercase",
							},
						},
						"standard_analyzer": map[string]interface{}{
							"tokenizer": "standard",
							"filter": []string{
								"lowercase",
							},
						},
					},
					"filter": map[string]interface{}{
						"autocomplete_filter": map[string]interface{}{
							"type":     "edge_ngram",
							"min_gram": 2,
							"max_gram": 20,
						},
					},
				},
			},
			"mappings": map[string]interface{}{
				"properties": properties,
			},
		}

		// Marshal alias body into JSON
		payload, err := json.Marshal(indexBody)
		if err != nil {
			logs.Error().Err(err).Msg("failed to marshal index body")
			response.SendErrorResponse(c, response.ErrResponse{
				Code:    http.StatusBadRequest,
				Message: "failed to marshal index body",
				Details: err.Error(),
				Type:    cErr.BadRequestError.String(),
			})
			return
		}

		ec := elastic.New(e)

		// Resolve the alias to the actual alias name
		indexName, err := ec.ResolveAlias(c.Request.Context(), alias)
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

			logs.Error().Err(err).Msg("failed to resolve alias")
			response.SendErrorResponse(c, response.ErrResponse{
				Code:    http.StatusInternalServerError,
				Message: "failed to resolve alias",
				Details: err.Error(),
				Type:    cErr.ServerError.String(),
			})
			return
		}

		// If the alias does not exist, create a new alias with the alias
		if indexName == "" {
			newIndexName := fmt.Sprintf("%s_%s", alias, time.Now().Format("20060102150405")) // Use timestamp as suffix
			if err := ec.CreateIndex(c.Request.Context(), newIndexName, payload); err != nil {
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
				logs.Error().Err(err).Msg("failed to create new index")
				c.JSON(http.StatusInternalServerError, response.ErrResponse{
					Code:    http.StatusInternalServerError,
					Message: "failed to create new index",
					Details: err.Error(),
					Type:    cErr.ServerError.String(),
				})
				return
			}

			// Attach the alias to the new alias
			if err := ec.UpdateAlias(c.Request.Context(), alias, newIndexName); err != nil {
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
				logs.Error().Err(err).Msg("failed to update alias")
				c.JSON(http.StatusInternalServerError, response.ErrResponse{
					Ts:      time.Now(),
					Code:    http.StatusInternalServerError,
					Message: "failed to update alias",
					Details: err.Error(),
					Type:    cErr.ServerError.String(),
				})
				return
			}

			c.JSON(http.StatusOK, response.SuccessResponse[any]{
				Ts:      time.Now(),
				Code:    http.StatusOK,
				Message: "index created/updated successfully",
			})
			return
		}

		// Attempt to update the index mappings
		if err := ec.UpdateIndexMappings(c.Request.Context(), indexName, payload); err != nil {
			logs.Warn().Err(err).Msg("failed to update index mappings")

			// If updating mappings fails, reindex the data
			newIndexName := fmt.Sprintf("%s_%s", alias, time.Now().Format("20060102150405")) // Use timestamp as suffix
			if err := ec.Reindex(c.Request.Context(), indexName, newIndexName, payload); err != nil {
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
				logs.Error().Err(err).Msg("failed to reindex data")
				response.SendErrorResponse(c, response.ErrResponse{
					Code:    http.StatusInternalServerError,
					Message: "failed to reindex data",
					Details: err.Error(),
					Type:    cErr.ServerError.String(),
				})
				return
			}

			// Update the alias to point to the new alias
			if err := ec.UpdateAlias(c.Request.Context(), alias, newIndexName); err != nil {
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
				logs.Error().Err(err).Msg("failed to update alias")
				response.SendErrorResponse(c, response.ErrResponse{
					Code:    http.StatusInternalServerError,
					Message: "failed to update alias",
					Details: err.Error(),
					Type:    cErr.ServerError.String(),
				})
				return
			}

			// Delete the old alias
			if err := ec.DeleteIndex(c.Request.Context(), indexName); err != nil {
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
				logs.Error().Err(err).Msg("failed to delete alias")
				response.SendErrorResponse(c, response.ErrResponse{
					Code:    http.StatusInternalServerError,
					Message: "failed to delete alias",
					Details: err.Error(),
					Type:    cErr.ServerError.String(),
				})
				return
			}

			c.JSON(http.StatusOK, response.SuccessResponse[any]{
				Ts:      time.Now(),
				Code:    http.StatusOK,
				Message: "Index updated successfully (re-indexed)",
			})
			return
		}
	}
}
