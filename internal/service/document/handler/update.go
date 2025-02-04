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
	"go-es/logger"
	"net/http"
	"time"
)

// UpdateDocument updates a document in an Elasticsearch index.
//
// This handler function extracts the index alias and document ID from the URI,
// parses the JSON request body to get update data, and calls the UpdateDocument
// function from the elastic package to perform the update. It sends an appropriate
// JSON response based on the result of the operation.
func UpdateDocument(e *elasticsearch.Client) func(c *gin.Context) {
	return func(c *gin.Context) {
		logs := logger.GetLogger(c)

		// Extract index name from the URI
		alias := c.Param("index")
		if alias == "" {
			logs.Error().Msg("alias is required in URI")
			response.SendErrorResponse(c, response.ErrResponse{
				Code:    http.StatusBadRequest,
				Message: "alias name is required in URI",
				Details: "alias name is required in URI",
				Type:    cErr.BadRequestError.String(),
			})
			return
		}

		// Extract document ID from the URI
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

		// Parse the JSON body
		var req request.UpdateDocumentRequest
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

		// Build the update body
		updateBody, err := json.Marshal(map[string]interface{}{
			"doc": req.Data, // Wrap the data in a "doc" field for partial updates
		})
		if err != nil {
			logs.Error().Err(err).Msg("failed to serialize update payload")
			response.SendErrorResponse(c, response.ErrResponse{
				Code:    http.StatusInternalServerError,
				Message: "internal server error",
				Details: err.Error(),
				Type:    cErr.ServerError.String(),
			})
			return
		}

		// Call UpdateDocument function
		ec := elastic.New(e)
		if err := ec.UpdateDocument(c.Request.Context(), alias, docID, updateBody); err != nil {
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
			logs.Error().Err(err).Msg("failed to update document")
			response.SendErrorResponse(c, response.ErrResponse{
				Code:    http.StatusInternalServerError,
				Message: "failed to update document",
				Details: err.Error(),
				Type:    cErr.ServerError.String(),
			})
			return
		}

		logs.Info().Msg("document updated successfully")
		c.JSON(http.StatusOK, response.SuccessResponse[any]{
			Ts:      time.Now(),
			Code:    http.StatusOK,
			Message: "document updated successfully",
		})
	}
}
