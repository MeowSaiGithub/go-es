package handler

import (
	"errors"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/gin-gonic/gin"
	cErr "go-es/internal/errors"
	"go-es/internal/response"
	"go-es/internal/service/index/elastic"
	"go-es/logger"
	"net/http"
	"time"
)

// DeleteIndex is a gin handler for deleting an index by its alias.
// It will resolve the alias to the actual index name and then delete the index.
// If the request fails, it will return an appropriate error response.
func DeleteIndex(e *elasticsearch.Client) func(*gin.Context) {
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

		// Delete the index
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

			logs.Error().Err(err).Msg("failed to delete index")
			response.SendErrorResponse(c, response.ErrResponse{
				Code:    http.StatusInternalServerError,
				Message: "failed to delete index",
				Details: err.Error(),
				Type:    cErr.ServerError.String(),
			})
			return
		}

		c.JSON(http.StatusOK, response.SuccessResponse[any]{
			Ts:      time.Now(),
			Code:    http.StatusOK,
			Message: "alias deleted successfully",
		})
		return
	}
}
