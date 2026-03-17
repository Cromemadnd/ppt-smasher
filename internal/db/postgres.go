package db

import (
	"context"

	"ppt-smasher/internal/config"

	"github.com/Wood-Q/Eino-pgvector/indexer"
	"github.com/Wood-Q/Eino-pgvector/retriever"
	"github.com/cloudwego/eino/components/embedding"
	einoindexer "github.com/cloudwego/eino/components/indexer"
	einoretriever "github.com/cloudwego/eino/components/retriever"
	"github.com/cloudwego/eino/schema"
)

type PostgresVectorDB struct {
	idx *indexer.Indexer
	ret *retriever.Retriever
}

func NewPostgresVectorDB(ctx context.Context, cfg *config.PostgresConfig, embedder embedding.Embedder, tableName string, dimension int) *PostgresVectorDB {
	// 初始化 Indexer
	idx, err := indexer.NewIndexer(ctx, &indexer.IndexerConfig{
		Host:       cfg.Host,
		Port:       cfg.Port,
		User:       cfg.User,
		Password:   cfg.Password,
		DBName:     cfg.DBName,
		SSLMode:    cfg.SSLMode,
		TableName:  tableName,
		Dimension:  dimension,
		Embedding:  embedder,
		VectorType: indexer.VectorTypeHalfvec,
		IndexType:  indexer.IndexTypeIVFFlat,
	})
	if err != nil {
		panic(err)
	}

	// 初始化 Retriever
	ret, err := retriever.NewRetriever(ctx, &retriever.RetrieverConfig{
		Host:      cfg.Host,
		Port:      cfg.Port,
		User:      cfg.User,
		Password:  cfg.Password,
		DBName:    cfg.DBName,
		SSLMode:   cfg.SSLMode,
		TableName: tableName,
		Dimension: dimension,
		Embedding: embedder,
	})
	if err != nil {
		panic(err)
	}

	return &PostgresVectorDB{
		idx: idx,
		ret: ret,
	}
}

func (v *PostgresVectorDB) Index(ctx context.Context, docs []*schema.Document, opts ...einoindexer.Option) ([]string, error) {
	// Ignore standard eino options for now as the plugin uses its own logic in Store
	return v.idx.Store(ctx, docs)
}

func (v *PostgresVectorDB) Retrieve(ctx context.Context, query string, opts ...einoretriever.Option) ([]*schema.Document, error) {
	// The plugin's Retrieve expects its own SearchOptions.
	searchOpts := &retriever.SearchOptions{
		Limit: 5, // Default limit
	}

	results, err := v.ret.Retrieve(ctx, query, searchOpts)
	if err != nil {
		return nil, err
	}

	docs := make([]*schema.Document, len(results))
	for i, r := range results {
		docs[i] = &schema.Document{
			ID:       r.ID,
			Content:  r.Content,
			MetaData: r.Metadata,
		}
	}

	return docs, nil
}

func (v *PostgresVectorDB) Close(ctx context.Context) error {
	return v.idx.Close()
}
