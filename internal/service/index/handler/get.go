package handler

import (
	"errors"
	"fmt"
	cErr "go-es/internal/errors"
	"go-es/internal/response"
	"go-es/logger"
	"net/http"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/gin-gonic/gin"
	"go-es/internal/service/index/elastic"
	"go-es/internal/service/index/model"
)

// GetIndex retrieves details of a single index in a structured format.
// It is a gin handler function that extracts the alias from the URI,
// resolves the alias to the actual index name, retrieves index details,
// and returns the index data in a JSON response.
func GetIndex(e *elasticsearch.Client) func(*gin.Context) {
	return func(c *gin.Context) {
		// Get logger for the current context
		logs := logger.GetLogger(c)

		// Extract alias from URI
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

		// Initialize elastic client
		ec := elastic.New(e)

		// Get index details using the alias
		info, err := ec.GetIndex(c.Request.Context(), alias)
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
			logs.Error().Err(err).Msg("failed to get index")
			response.SendErrorResponse(c, response.ErrResponse{
				Code:    http.StatusInternalServerError,
				Message: "failed to get index",
				Details: err.Error(),
				Type:    cErr.ServerError.String(),
			})
			return
		}

		// Parse the raw Elasticsearch response into the desired format
		indexData, err := parseIndex(info, alias)
		if err != nil {
			logs.Error().Err(err).Msg("failed to parse index data")
			response.SendErrorResponse(c, response.ErrResponse{
				Code:    http.StatusInternalServerError,
				Message: "failed to parse index data",
				Details: err.Error(),
				Type:    cErr.ParseError.String(),
			})
			return
		}

		// Send success response with index data
		c.JSON(http.StatusOK, response.SuccessResponse[model.Index]{
			Ts:      time.Now(),
			Code:    http.StatusOK,
			Message: "index info retrieved successfully",
			Data:    *indexData,
		})
	}
}

// parseIndex parses the raw Elasticsearch response into the desired format.
func parseIndex(info map[string]interface{}, index string) (*model.Index, error) {
	// Extract the index map
	indexMap, ok := info[index].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid index map format for %s", index)
	}

	// Extract mappings
	mappings, ok := indexMap["mappings"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("failed to extract mappings for %s", index)
	}

	// Extract properties
	properties, ok := mappings["properties"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("failed to extract properties for %s", index)
	}

	// Parse fields
	fields, err := parseProperties(properties)
	if err != nil {
		return nil, fmt.Errorf("failed to parse properties for %s: %w", index, err)
	}

	return &model.Index{
		Name:   index,
		Fields: fields,
	}, nil
}

// parseProperties parses field properties into FieldConfig format.
func parseProperties(properties map[string]interface{}) (map[string]model.FieldConfig, error) {
	fields := make(map[string]model.FieldConfig)

	for fieldName, fieldInfo := range properties {
		field, ok := fieldInfo.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid field format for %s", fieldName)
		}

		// Extract type and optional attributes
		fieldType, _ := field["type"].(string)
		analyzer, _ := field["analyzer"].(string)
		searchAnalyzer, _ := field["search_analyzer"].(string)

		// Determine autocomplete and search flags
		autocomplete := analyzer == "edge_ngram_analyzer"
		search := searchAnalyzer == "standard" || analyzer == "edge_ngram_analyzer"

		// Populate FieldConfig
		fields[fieldName] = model.FieldConfig{
			Type:           fieldType,
			Analyzer:       analyzer,
			SearchAnalyzer: searchAnalyzer,
			Autocomplete:   autocomplete,
			Search:         search,
		}
	}

	return fields, nil
}
