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

// Exists checks if an index associated with an alias exists in Elasticsearch.
//
// It retrieves the alias from the URI, resolves it to the actual index name,
// and checks if the index exists. It returns a JSON response indicating the
// existence of the alias or an appropriate error message if the operation fails.
func Exists(e *elasticsearch.Client) func(*gin.Context) {
	return func(c *gin.Context) {
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

		ec := elastic.New(e)

		// Resolve the alias to the actual index name
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

		// Check if the index exists
		_, err = ec.CheckIndex(c.Request.Context(), indexName)
		if err != nil {
			var esErr cErr.ElasticsearchError
			if errors.As(err, &esErr) {
				logs.Error().Err(esErr.Details).Msg(esErr.Message)
				c.JSON(esErr.StatusCode, response.ErrResponse{
					Ts:      time.Now(),
					Code:    esErr.StatusCode,
					Message: esErr.Message,
					Details: esErr.Details.Error(),
					Type:    esErr.Type.String(),
				})
				return
			}

			logs.Error().Err(err).Msg("failed to check index existence")
			response.SendErrorResponse(c, response.ErrResponse{
				Code:    http.StatusInternalServerError,
				Message: "failed to check index existence",
				Details: err.Error(),
				Type:    cErr.ServerError.String(),
			})
			return
		}

		// Respond with success if the index exists
		c.JSON(http.StatusOK, response.SuccessResponse[any]{
			Ts:      time.Now(),
			Code:    http.StatusOK,
			Message: "alias exists",
		})
	}
}
