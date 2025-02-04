package handler

import (
	"errors"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/gin-gonic/gin"
	cErr "go-es/internal/errors"
	"go-es/internal/response"
	"go-es/internal/service/document/elastic"
	"go-es/internal/service/document/model/request"
	"go-es/logger"
	"net/http"
	"time"
)

// AddData adds documents to an Elasticsearch index.
// It extracts the alias from the URI, parses the JSON request body,
// validates the data, and then calls AddBatchData to add the documents.
func AddData(e *elasticsearch.Client) func(c *gin.Context) {
	return func(c *gin.Context) {
		logs := logger.GetLogger(c)

		// Extract alias name from the URI
		alias := c.Param("alias")
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

		// Parse the JSON body into the AddDataRequest structure
		var req request.AddDataRequest
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

		// Validate that there is data to upload
		if len(req.Data) == 0 {
			logs.Error().Msg("data array cannot be empty")
			response.SendErrorResponse(c, response.ErrResponse{
				Code:    http.StatusBadRequest,
				Message: "data array cannot be empty",
				Details: "data array cannot be empty",
				Type:    cErr.BadRequestError.String(),
			})
			return
		}

		// Call AddBatchData to add the documents to Elasticsearch
		ec := elastic.New(e)
		err := ec.AddBatchData(c.Request.Context(), alias, req.Data)
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
			logs.Error().Err(err).Msg("failed to add batch data")
			response.SendErrorResponse(c, response.ErrResponse{
				Code:    http.StatusInternalServerError,
				Message: "failed to add batch data",
				Details: err.Error(),
				Type:    cErr.ServerError.String(),
			})
			return
		}

		// Respond with success
		logs.Info().Msg("data added successfully")
		c.JSON(http.StatusOK, response.SuccessResponse[any]{
			Ts:      time.Now(),
			Code:    http.StatusOK,
			Message: "data added successfully",
		})
	}
}
