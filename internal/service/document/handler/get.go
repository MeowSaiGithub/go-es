package handler

import (
	"errors"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/gin-gonic/gin"
	cErr "go-es/internal/errors"
	"go-es/internal/response"
	"go-es/internal/service/document/elastic"
	"go-es/logger"
	"net/http"
	"time"
)

// GetDocumentByID retrieves a document by its ID from an Elasticsearch index.
//
// This function validates the alias and document ID from the URI, retrieves
// the document from Elasticsearch, and returns it in the response. If any
// errors occur, it sends an appropriate error response.
func GetDocumentByID(e *elasticsearch.Client) func(c *gin.Context) {
	return func(c *gin.Context) {
		logs := logger.GetLogger(c)

		// Validate alias from URI
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

		// Validate document ID from URI
		docID := c.Param("id")
		if docID == "" {
			logs.Error().Msg("document ID is required")
			response.SendErrorResponse(c, response.ErrResponse{
				Code:    http.StatusBadRequest,
				Message: "document ID is required",
				Details: "document ID is required",
				Type:    cErr.BadRequestError.String(),
			})
			return
		}

		// Resolve the alias to the actual index name and get the document
		ec := elastic.New(e)
		doc, err := ec.GetDocumentByID(c.Request.Context(), alias, docID)
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
			logs.Error().Err(err).Msg("failed to get document by id")
			response.SendErrorResponse(c, response.ErrResponse{
				Code:    http.StatusInternalServerError,
				Message: "failed to get document by id",
				Details: err.Error(),
				Type:    cErr.ServerError.String(),
			})
			return
		}

		// Return the document in the response
		c.JSON(http.StatusOK, response.SuccessResponse[map[string]interface{}]{
			Ts:      time.Now(),
			Code:    http.StatusOK,
			Message: "document fetched successfully",
			Data:    gin.H{"document": doc["_source"]},
		})
	}
}
