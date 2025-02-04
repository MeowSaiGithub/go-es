package elastic

import (
	"context"
	"github.com/elastic/go-elasticsearch/v8"
	"go-es/internal/service/document/model/response"
)

// Elastic is an interface to interact with Elasticsearch
type Elastic interface {
	// AddBatchData adds a batch of documents to an Elasticsearch index
	AddBatchData(ctx context.Context, indexName string, data []map[string]interface{}) error
	// Search runs a search query against an Elasticsearch index
	Search(ctx context.Context, indexName string, payload []byte) (*response.SearchResponse, error)
	// Suggest runs a suggest query against an Elasticsearch index
	AutoComplete(ctx context.Context, indexName string, payload []byte) (*response.SuggestResponse, error)
	// DeleteDocument deletes a document from an Elasticsearch index
	DeleteDocument(ctx context.Context, index string, docId string) error
	// UpdateDocument updates a document in an Elasticsearch index
	UpdateDocument(ctx context.Context, index string, docID string, body []byte) error
	// GetDocumentByID retrieves a document by its ID
	GetDocumentByID(ctx context.Context, index string, docID string) (map[string]interface{}, error)
	// ListAllDocuments retrieves all documents from an Elasticsearch index
	ListAllDocuments(ctx context.Context, index string, query map[string]interface{}) (map[string]interface{}, error)
	// ExportDocuments exports documents from an Elasticsearch index
	ExportDocuments(ctx context.Context, index string, query map[string]interface{}) ([]map[string]interface{}, error)
	// ImportDocuments imports documents to an Elasticsearch index
	ImportDocuments(ctx context.Context, index string, documents []map[string]interface{}) error
}
type elastic struct {
	client *elasticsearch.Client
}

// New creates a new instance of Elastic.
func New(client *elasticsearch.Client) Elastic {
	return &elastic{client: client}
}
