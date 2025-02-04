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
	"go-es/internal/service/index/model"
	"go-es/logger"
	"net/http"
	"time"
)

// CreateIndex handles creating a new index in Elasticsearch with the specified name and settings.
//
// It sends an index creation request to the Elasticsearch server using the provided index name
// and payload, and returns an error if the operation fails.
func CreateIndex(client *elasticsearch.Client) func(c *gin.Context) {
	return func(c *gin.Context) {
		logs := logger.GetLogger(c)

		var req model.Index
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

		// Generate properties for the index
		properties, err := generateProperties(req.Fields)
		if err != nil {
			logs.Error().Err(err).Msg("failed to generate fields")
			response.SendErrorResponse(c, response.ErrResponse{
				Code:    http.StatusBadRequest,
				Message: "failed to generate properties fields",
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
						"autocomplete_analyzer": map[string]interface{}{ //  FIX: Change tokenizer
							"tokenizer": "edge_ngram", // Use "edge_ngram" for prefix-based suggestions
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
							"type":     "edge_ngram", //  Ensures prefix search works
							"min_gram": 2,
							"max_gram": 20,
						},
					},
				},
			},
			"mappings": map[string]interface{}{
				"properties": properties,
			},
			"aliases": map[string]interface{}{
				req.Name: map[string]interface{}{},
			},
		}

		// Marshal index body into JSON
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

		indexName := fmt.Sprintf("%s_%s", req.Name, time.Now().Format("20060102150405"))

		ec := elastic.New(client)

		// Resolve the alias to the actual alias name
		_, err = ec.ResolveAlias(c.Request.Context(), req.Name)
		if err != nil {
			var esErr cErr.ElasticsearchError
			if errors.As(err, &esErr) {
				if esErr.Type != cErr.NotFoundError {
					logs.Error().Err(esErr.Details).Msg(esErr.Message)
					response.SendErrorResponse(c, response.ErrResponse{
						Code:    esErr.StatusCode,
						Message: esErr.Message,
						Details: esErr.Details.Error(),
						Type:    esErr.Type.String(),
					})
					return
				}
			} else {
				logs.Error().Err(err).Msg("failed to resolve alias")
				response.SendErrorResponse(c, response.ErrResponse{
					Code:    http.StatusInternalServerError,
					Message: "failed to resolve alias",
					Details: err.Error(),
					Type:    cErr.ServerError.String(),
				})
				return
			}
		} else {
			logs.Error().Str("alias", req.Name).Msg("index/alias already exists")
			response.SendErrorResponse(c, response.ErrResponse{
				Code:    http.StatusConflict,
				Message: "index/alias already exists",
				Details: "index/alias already exists",
				Type:    cErr.ResourceAlreadyExistsException.String(),
			})
			return
		}

		if err := ec.CreateIndex(c.Request.Context(), indexName, payload); err != nil {
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
			logs.Error().Err(err).Msg("failed to create index")
			response.SendErrorResponse(c, response.ErrResponse{
				Code:    http.StatusInternalServerError,
				Message: "failed to create index",
				Details: err.Error(),
				Type:    cErr.ServerError.String(),
			})
			return
		}

		c.JSON(http.StatusOK, response.SuccessResponse[any]{
			Ts:      time.Now(),
			Code:    http.StatusOK,
			Message: "index created successfully",
		})
	}
}

// generateProperties generates the properties mapping for the given field configurations
// and validates the results. It returns an error if any of the fields have invalid
// configurations.
func generateProperties(fields map[string]model.FieldConfig) (map[string]interface{}, error) {
	properties := make(map[string]interface{})
	// Iterate over the fields and generate their mappings

	for fieldName, fieldConfig := range fields {
		fieldMapping := map[string]interface{}{
			"type": fieldConfig.Type,
		}

		// Handle autocomplete & search fields (Only for `text`)
		if fieldConfig.Type == "text" && (fieldConfig.Autocomplete || fieldConfig.Search) {
			fieldMapping["fields"] = map[string]interface{}{
				"raw": map[string]interface{}{
					"type": "keyword",
				},
			}

			// Add proper `completion` field for suggestions
			if fieldConfig.Autocomplete {
				fieldMapping["fields"].(map[string]interface{})["suggest"] = map[string]interface{}{
					"type": "completion",
				}
			}

			// Add full-text search field
			if fieldConfig.Search {
				fieldMapping["fields"].(map[string]interface{})["fulltext"] = map[string]interface{}{
					"type":     "text",
					"analyzer": "standard_analyzer",
				}
			}
		}

		// Handle nested fields
		if fieldConfig.Type == "nested" && len(fieldConfig.Properties) > 0 {
			nestedProperties, err := generateProperties(fieldConfig.Properties)
			if err != nil {
				return nil, err
			}
			fieldMapping["properties"] = nestedProperties
		}

		properties[fieldName] = fieldMapping
	}

	return properties, nil
}
