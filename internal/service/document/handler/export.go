package handler

import (
	"errors"
	"fmt"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/gin-gonic/gin"
	"github.com/goccy/go-json"
	cErr "go-es/internal/errors"
	"go-es/internal/response"
	"go-es/internal/service/document/elastic"
	"go-es/internal/service/document/model/request"
	"go-es/logger"
	"net/http"
	"strings"
)

// ExportDocuments returns documents in an Elasticsearch index
//
// It takes the alias name as a parameter from the URI and uses it to query
// the Elasticsearch cluster for documents. It returns a JSON response containing
// the documents. If the "bulk" query parameter is set to true, it returns the
// documents in bulk API format. Otherwise, it returns the documents in standard
// JSON format.
func ExportDocuments(e *elasticsearch.Client) func(c *gin.Context) {
	return func(c *gin.Context) {
		logs := logger.GetLogger(c)

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

		// Parse the JSON body for filters (optional)
		var req request.ExportDocumentsRequest
		if c.Request.ContentLength > 0 { // Only parse body if it exists
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
		}

		// Determine if the output should be in bulk API format
		bulk := c.DefaultQuery("bulk", "false") // e.g., "true" or "false"
		isBulk := bulk == "true"

		// Call ExportDocuments function
		ec := elastic.New(e)
		documents, err := ec.ExportDocuments(c.Request.Context(), alias, req.Query)
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
			logs.Error().Err(err).Msg("failed to export documents")
			response.SendErrorResponse(c, response.ErrResponse{
				Code:    http.StatusInternalServerError,
				Message: "failed to export documents",
				Details: err.Error(),
				Type:    cErr.ServerError.String(),
			})
			return
		}

		// Set response headers
		if isBulk {
			// Bulk format → NDJSON
			c.Header("Content-Type", "application/x-ndjson")
			c.Header("Content-Disposition", "attachment; filename=export.ndjson")

			// Convert documents to bulk API NDJSON format
			var bulkPayload strings.Builder
			for _, doc := range documents {
				// Add the action metadata line
				bulkPayload.WriteString(fmt.Sprintf(`{"index":{"_index":"%s"}}`+"\n", alias))
				// Add the document data
				docBytes, _ := json.Marshal(doc)
				bulkPayload.WriteString(string(docBytes) + "\n")
			}

			// Return NDJSON data as response
			c.String(http.StatusOK, bulkPayload.String())

		} else {
			// Standard JSON format
			c.Header("Content-Type", "application/json")
			c.Header("Content-Disposition", "attachment; filename=export.json")
			c.JSON(http.StatusOK, documents)
		}

	}
}
