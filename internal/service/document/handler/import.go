package handler

import (
	"bufio"
	"bytes"
	"errors"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/gin-gonic/gin"
	"github.com/goccy/go-json"
	cErr "go-es/internal/errors"
	"go-es/internal/response"
	"go-es/internal/service/document/elastic"
	"go-es/internal/service/document/model/request"
	"go-es/logger"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"
	"time"
)

// ImportDocuments handles importing documents into Elasticsearch.
//
// This handler function parses the incoming request to determine the source of
// the documents to be imported, either from an uploaded file or a JSON string.
// It supports bulk import if specified, and uses the ImportDocuments method
// from the elastic package to insert the documents into the specified index.
func ImportDocuments(e *elasticsearch.Client) func(*gin.Context) {
	return func(c *gin.Context) {
		logs := logger.GetLogger(c)

		// Bind the incoming request to the ImportRequest structure
		var req request.ImportRequest
		if err := c.ShouldBind(&req); err != nil {
			logs.Error().Err(err).Msg("invalid request payload")
			response.SendErrorResponse(c, response.ErrResponse{
				Code:    http.StatusBadRequest,
				Message: "invalid JSON payload",
				Details: err.Error(),
				Type:    cErr.BadRequestError.String(),
			})
			return
		}

		// Extract and parse the `bulk` query parameter if not provided in the form
		if !req.Bulk {
			bulkParam := c.DefaultQuery("bulk", "false")
			req.Bulk, _ = strconv.ParseBool(bulkParam)
		}

		var documents []map[string]interface{}
		var err error

		// Determine the source of input, file or JSON content
		if req.File != nil {
			documents, err = parseFile(req.File, req.Bulk)
		} else if req.JSON != "" {
			documents, err = parseJSON(req.JSON, req.Bulk)
		} else {
			logs.Error().Msg("no file or JSON content provided")
			response.SendErrorResponse(c, response.ErrResponse{
				Code:    http.StatusBadRequest,
				Message: "no file or JSON content provided",
				Details: "no file or JSON content provided",
				Type:    cErr.BadRequestError.String(),
			})
			return
		}

		// Handle error if parsing inputs fails
		if err != nil {
			logs.Error().Err(err).Msg("failed to parse inputs")
			response.SendErrorResponse(c, response.ErrResponse{
				Code:    http.StatusBadRequest,
				Message: "failed to parse input",
				Details: err.Error(),
				Type:    cErr.BadRequestError.String(),
			})
			return
		}

		// Attempt to import documents into Elasticsearch
		ec := elastic.New(e)
		if err := ec.ImportDocuments(c.Request.Context(), req.Index, documents); err != nil {
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
			logs.Error().Err(err).Msg("failed to import documents")
			response.SendErrorResponse(c, response.ErrResponse{
				Code:    http.StatusInternalServerError,
				Message: "failed to import documents",
				Details: err.Error(),
				Type:    cErr.ServerError.String(),
			})
			return
		}

		// Success response
		logs.Info().Msg("documents imported successfully")
		c.JSON(http.StatusOK, response.SuccessResponse[any]{
			Ts:      time.Now(),
			Code:    http.StatusOK,
			Message: "documents imported successfully",
		})
	}
}

// parseFile reads and parses the uploaded file.
//
// The uploaded file is read and parsed into individual documents based on the
// bulk flag. If the bulk flag is set to true, the file is expected to be in
// Elasticsearch bulk format. Otherwise, each line of the file is expected to
// contain a single JSON document.
func parseFile(fileHeader *multipart.FileHeader, isBulk bool) ([]map[string]interface{}, error) {
	file, err := fileHeader.Open()
	if err != nil {
		return nil, errors.New("failed to open import file")
	}
	defer file.Close()

	// Read the file content
	fileContent, err := io.ReadAll(file)
	if err != nil {
		return nil, errors.New("failed to read file content")
	}

	// Parse based on the bulk flag
	if isBulk {
		return parseBulkJSON(fileContent)
	}
	return parsePureJSON(fileContent)
}

// parseJSON parses the JSON content provided in the request body.
//
// It accepts a JSON string and a boolean indicating whether the content
// is in bulk format. If the bulk flag is true, the function delegates
// the parsing to parseBulkJSON, which handles Elasticsearch bulk format.
// Otherwise, it uses parsePureJSON to parse the JSON as a pure array.
func parseJSON(jsonStr string, isBulk bool) ([]map[string]interface{}, error) {
	// Convert the JSON string to a byte slice
	jsonBytes := []byte(jsonStr)

	// Delegate parsing based on the bulk flag
	if isBulk {
		return parseBulkJSON(jsonBytes)
	}
	return parsePureJSON(jsonBytes)
}

// parseBulkJSON parses JSON in Elasticsearch bulk format.
//
// This function takes a byte slice of data in bulk format, where each
// document is preceded by a metadata line. It scans through the data,
// verifying the presence of an "index" field in the metadata and
// unmarshals both metadata and document lines into a slice of maps.
// It returns the documents or an error if parsing fails.
func parseBulkJSON(data []byte) ([]map[string]interface{}, error) {
	var documents []map[string]interface{}
	scanner := bufio.NewScanner(bytes.NewReader(data))

	for scanner.Scan() {
		// Read the metadata line
		metaLine := scanner.Bytes()
		var meta map[string]interface{}
		if err := json.Unmarshal(metaLine, &meta); err != nil {
			return nil, errors.New("invalid metadata line in bulk JSON")
		}

		// Ensure the metadata line contains an "index" field
		if _, exists := meta["index"]; !exists {
			return nil, errors.New("metadata line must contain an 'index' field")
		}

		// Read the document line
		if !scanner.Scan() {
			return nil, errors.New("missing document line in bulk JSON")
		}
		docLine := scanner.Bytes()
		var doc map[string]interface{}
		if err := json.Unmarshal(docLine, &doc); err != nil {
			return nil, errors.New("invalid document line in bulk JSON")
		}

		// Append the document to the list
		documents = append(documents, doc)
	}

	if err := scanner.Err(); err != nil {
		return nil, errors.New("failed to read bulk JSON")
	}

	return documents, nil
}

// parsePureJSON parses JSON in pure array format.
// It takes a byte slice and unmarshals it into a slice of maps.
// It returns the documents or an error if parsing fails.
func parsePureJSON(data []byte) ([]map[string]interface{}, error) {
	var documents []map[string]interface{}
	if err := json.Unmarshal(data, &documents); err != nil {
		return nil, errors.New("invalid JSON format")
	}
	return documents, nil
}
