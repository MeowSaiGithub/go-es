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

// DeleteDocument handles deleting a document in an Elasticsearch index.
//
// This function will validate the inputs (alias and document ID), resolve the
// alias to the actual index name, and then delete the document.
func DeleteDocument(e *elasticsearch.Client) func(c *gin.Context) {
	return func(c *gin.Context) {
		logs := logger.GetLogger(c)

		// Validate the inputs
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

		docID := c.Param("id")
		if docID == "" {
			logs.Error().Msg("document ID are required")
			response.SendErrorResponse(c, response.ErrResponse{
				Code:    http.StatusBadRequest,
				Message: "document ID are required",
				Details: "document ID are required",
				Type:    cErr.BadRequestError.String(),
			})
			return
		}

		// Resolve the alias to the actual index name
		ec := elastic.New(e)
		err := ec.DeleteDocument(c, alias, docID)
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
			logs.Error().Err(err).Msg("failed to delete document")
			response.SendErrorResponse(c, response.ErrResponse{
				Code:    http.StatusInternalServerError,
				Message: "failed to delete document",
				Details: err.Error(),
				Type:    cErr.ServerError.String(),
			})
			return
		}

		logs.Info().Str("doc_id", docID).Msg("document deleted")
		// Return a success response
		c.JSON(http.StatusOK, response.SuccessResponse[any]{
			Ts:      time.Now(),
			Code:    http.StatusOK,
			Message: "document deleted",
		})
	}
}
