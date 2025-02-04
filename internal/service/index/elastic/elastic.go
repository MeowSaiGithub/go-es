package elastic

import (
	"context"
	"github.com/elastic/go-elasticsearch/v8"
)

// Elastic is an interface to interact with Elasticsearch
type Elastic interface {
	// CheckIndex checks if an index exists on Elasticsearch.
	CheckIndex(ctx context.Context, index string) (bool, error)
	// CreateIndex creates an index on Elasticsearch.
	CreateIndex(ctx context.Context, index string, payload []byte) error
	// DeleteIndex deletes an index on Elasticsearch.
	DeleteIndex(ctx context.Context, index string) error
	// GetIndex retrieves the settings and mappings of an index on Elasticsearch.
	GetIndex(ctx context.Context, index string) (map[string]interface{}, error)
	// ListIndices retrieves a list of all indices on Elasticsearch.
	ListIndices(ctx context.Context) (map[string]string, error)
	// UpdateIndexMappings updates the mappings of an existing index.
	UpdateIndexMappings(ctx context.Context, index string, payload []byte) error
	// Reindex copies data from the old index to a new index with updated mappings.
	Reindex(ctx context.Context, oldIndex, newIndex string, payload []byte) error
	// UpdateAlias updates the alias of an index on Elasticsearch.
	UpdateAlias(ctx context.Context, alias, newIndex string) error
	// ResolveAlias resolves the index of an alias on Elasticsearch.
	ResolveAlias(ctx context.Context, alias string) (string, error)
}
type elastic struct {
	client *elasticsearch.Client
}

func New(client *elasticsearch.Client) Elastic {
	return &elastic{client: client}
}
