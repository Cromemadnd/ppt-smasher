package db

import (
	"context"
	"fmt"
	"log"

	"ppt-smasher/internal/config"

	"github.com/cloudwego/eino-ext/components/embedding/ark"
	"github.com/cloudwego/eino-ext/components/embedding/openai"
	"github.com/cloudwego/eino/components/embedding"
	"github.com/cloudwego/eino/schema"
)

var (
	embedder embedding.Embedder
	vectorDB *PostgresVectorDB
)

func InitVectorDB(ctx context.Context) {
	emb := getEmbedder(ctx)
	if emb == nil {
		log.Fatalf("failed to get embedder for vector db initialization")
	}

	conf := config.GlobalConfig.LLM
	dim := conf.EmbeddingDim
	if dim == 0 {
		dim = 384
	}

	vectorDB = NewPostgresVectorDB(ctx, &config.GlobalConfig.Postgres, emb, "document_chunks", dim)
	log.Println("Postgres VectorDB initialized successfully.")
}

func getEmbedder(ctx context.Context) embedding.Embedder {
	if embedder != nil {
		return embedder
	}
	conf := config.GlobalConfig.LLM

	dim := conf.EmbeddingDim
	if dim == 0 {
		dim = 384 // default
	}
	model := conf.EmbeddingModel
	if model == "" {
		model = "text-embedding-3-small"
	}

	if conf.APIKey == "" {
		log.Println("WARNING: no API key provided for embedder")
	}

	switch conf.EmbeddingProvider {
	case "openai", "":
		emb, err := openai.NewEmbedder(ctx, &openai.EmbeddingConfig{
			APIKey:     conf.APIKey,
			BaseURL:    conf.BaseURL,
			Model:      model,
			Dimensions: &dim,
		})
		if err != nil {
			log.Printf("failed to init openai embedder: %v", err)
			return nil
		}
		embedder = emb
	case "ark":
		emb, err := ark.NewEmbedder(ctx, &ark.EmbeddingConfig{
			APIKey: conf.APIKey,
			Model:  model,
		})
		if err != nil {
			log.Printf("failed to init ark embedder: %v", err)
			return nil
		}
		embedder = emb
	default:
		log.Printf("unsupported embedding provider: %s", conf.EmbeddingProvider)
		return nil
	}

	return embedder
}

func getEmbedding(ctx context.Context, text string) ([]float32, error) {
	emb := getEmbedder(ctx)
	if emb == nil {
		return nil, fmt.Errorf("embedder not available")
	}
	// eino EmbedStrings returns [][]float64 usually
	vectors, err := emb.EmbedStrings(ctx, []string{text})
	if err != nil {
		return nil, err
	}
	if len(vectors) == 0 {
		return nil, fmt.Errorf("no vector returned")
	}

	// Convert []float64 to []float32 for LanceDB
	v64 := vectors[0]
	v32 := make([]float32, len(v64))
	for i, v := range v64 {
		v32[i] = float32(v)
	}
	return v32, nil
}

// AddDocumentChunk 插入文档片段到 pgvector
func AddDocumentChunk(ctx context.Context, theme string, id string, text string) error {
	if vectorDB == nil {
		return fmt.Errorf("vectorDB not initialized")
	}

	_, err := vectorDB.Index(ctx, []*schema.Document{
		{
			ID:      id,
			Content: text,
			MetaData: map[string]interface{}{
				"theme": theme,
			},
		},
	})
	return err
}

// SearchDocument 检索相关的文档片段
func SearchDocument(ctx context.Context, theme string, query string, topK int) ([]string, error) {
	if vectorDB == nil {
		return nil, fmt.Errorf("vectorDB not initialized")
	}

	// NOTE: The current plugin implementation might not support metadata filtering in Retrieve directly
	// unless modified. Using it as is for now.
	docs, err := vectorDB.Retrieve(ctx, query)
	if err != nil {
		return nil, err
	}

	var chunks []string
	for _, doc := range docs {
		// Filter by theme if stored in metadata
		if t, ok := doc.MetaData["theme"].(string); ok && t == theme {
			chunks = append(chunks, doc.Content)
		} else if !ok {
			// If no theme metadata found (depends on plugin version), return all
			chunks = append(chunks, doc.Content)
		}
	}

	if len(chunks) > topK {
		chunks = chunks[:topK]
	}

	return chunks, nil
}
