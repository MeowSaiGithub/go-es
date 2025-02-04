package handler

import (
	"errors"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	cErr "go-es/internal/errors"
	"go-es/internal/response"
	"go-es/internal/service/document/elastic"
	docResp "go-es/internal/service/document/model/response"
	"net/http"
	"strconv"
	"time"
)

// ListAllDocuments lists all documents in an Elasticsearch index
//
// This endpoint is used to list all documents in an Elasticsearch index. It takes
// the alias name as a parameter from the URI and uses it to query the Elasticsearch
// cluster for documents. It returns a JSON response containing the documents.
//
// The endpoint also supports pagination. The pagination parameters are `page` and
// `size`. The `page` parameter specifies the page number and the `size` parameter
// specifies the number of documents per page. Both parameters are optional. If
// not provided, the endpoint will return the first page of documents with a size
// of 10.
func ListAllDocuments(e *elasticsearch.Client) func(ctx *gin.Context) {
	return func(c *gin.Context) {
		logs := log.With().Str("handler", "ListAllDocuments").Logger()

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

		// Parse pagination parameters
		page := c.DefaultQuery("page", "1")
		size := c.DefaultQuery("size", "10")

		pageInt, err := strconv.Atoi(page)
		if err != nil {
			logs.Error().Msg("invalid page number")
			response.SendErrorResponse(c, response.ErrResponse{
				Code:    http.StatusBadRequest,
				Message: "invalid page number",
				Details: "invalid page number",
				Type:    cErr.BadRequestError.String(),
			})
			return
		}

		sizeInt, err := strconv.Atoi(size)
		if err != nil {
			logs.Error().Msg("invalid page size")
			response.SendErrorResponse(c, response.ErrResponse{
				Code:    http.StatusBadRequest,
				Message: "invalid page size",
				Details: "invalid page size",
				Type:    cErr.BadRequestError.String(),
			})
			return
		}

		// Query all documents
		query := map[string]interface{}{
			"from": (pageInt - 1) * sizeInt,
			"size": sizeInt,
		}

		ec := elastic.New(e)
		searchResult, err := ec.ListAllDocuments(c.Request.Context(), alias, query)
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
			logs.Error().Err(err).Msg("failed to list all documents")
			response.SendErrorResponse(c, response.ErrResponse{
				Code:    http.StatusInternalServerError,
				Message: "failed to list all documents",
				Details: err.Error(),
				Type:    cErr.ServerError.String(),
			})
			return
		}

		// Transform the Elasticsearch resp into the simplified format
		resp := transformSearchResult(searchResult)

		logs.Info().Msg("documents listed successfully")
		c.JSON(http.StatusOK, response.SuccessResponse[docResp.ListDocumentsResponse]{
			Ts:      time.Now(),
			Code:    http.StatusOK,
			Message: "success",
			Data:    resp,
		})
	}
}

// transformSearchResult transforms the Elasticsearch response into the simplified format
func transformSearchResult(searchResult map[string]interface{}) docResp.ListDocumentsResponse {
	hits := searchResult["hits"].(map[string]interface{})["hits"].([]interface{})
	total := int(searchResult["hits"].(map[string]interface{})["total"].(map[string]interface{})["value"].(float64))

	documents := make([]docResp.Document, 0, len(hits))
	for _, hit := range hits {
		hitMap := hit.(map[string]interface{})
		documents = append(documents, docResp.Document{
			ID:   hitMap["_id"].(string),
			Data: hitMap["_source"].(map[string]interface{}),
		})
	}

	return docResp.ListDocumentsResponse{
		Total:     total,
		Documents: documents,
	}
}
