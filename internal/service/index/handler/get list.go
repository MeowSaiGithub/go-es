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

// ListIndices returns a list of indices in Elasticsearch.
// It calls the Elasticsearch _cat/indices API to get a list of all indices,
// and the _cat/aliases API to get a list of all aliases. It then filters out
// indices starting with '.' (system/internal indices), and returns a map of
// index names to their corresponding aliases.
func ListIndices(e *elasticsearch.Client) func(*gin.Context) {
	return func(c *gin.Context) {
		logs := logger.GetLogger(c)

		ec := elastic.New(e)

		indices, err := ec.ListIndices(c.Request.Context())
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
			logs.Error().Err(err).Msg("failed to list indices")
			response.SendErrorResponse(c, response.ErrResponse{
				Code:    http.StatusInternalServerError,
				Message: "failed to list indices",
				Details: err.Error(),
				Type:    cErr.ServerError.String(),
			})
			return
		}

		c.JSON(http.StatusOK, response.SuccessResponse[map[string]string]{
			Ts:      time.Now(),
			Code:    http.StatusOK,
			Message: "success",
			Data:    indices,
		})
		return
	}
}
