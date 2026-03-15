package db

import (
	"context"
	"fmt"
	"log"

	"ppt-stasher-backend/internal/config"

	"github.com/apache/arrow/go/v17/arrow"
	"github.com/apache/arrow/go/v17/arrow/array"
	"github.com/apache/arrow/go/v17/arrow/memory"
	"github.com/cloudwego/eino-ext/components/embedding/ark"
	"github.com/cloudwego/eino-ext/components/embedding/openai"
	"github.com/cloudwego/eino/components/embedding"
)

var embedder embedding.Embedder

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

// AddDocumentChunk 插入文档片段到 LanceDB
func AddDocumentChunk(ctx context.Context, theme string, id string, text string) error {
	if LanceTable == nil {
		return fmt.Errorf("LanceTable not initialized")
	}

	vectorData, err := getEmbedding(ctx, text)
	if err != nil {
		return fmt.Errorf("failed to embed document chunk: %w", err)
	}
	pool := memory.NewGoAllocator()

	// 1. ID builder
	idBuilder := array.NewStringBuilder(pool)
	idBuilder.AppendValues([]string{id}, nil)
	idArray := idBuilder.NewArray()
	defer idArray.Release()

	// 2. Theme builder
	themeBuilder := array.NewStringBuilder(pool)
	themeBuilder.AppendValues([]string{theme}, nil)
	themeArray := themeBuilder.NewArray()
	defer themeArray.Release()

	// 3. Text builder
	textBuilder := array.NewStringBuilder(pool)
	textBuilder.AppendValues([]string{text}, nil)
	textArray := textBuilder.NewArray()
	defer textArray.Release()

	// 4. Vector builder
	vectorFloat32Builder := array.NewFloat32Builder(pool)
	vectorFloat32Builder.AppendValues(vectorData, nil)
	vectorFloat32Array := vectorFloat32Builder.NewArray()
	defer vectorFloat32Array.Release()

	dim := int32(config.GlobalConfig.LLM.EmbeddingDim)
	if dim == 0 {
		dim = 384
	}
	vectorListType := arrow.FixedSizeListOf(dim, arrow.PrimitiveTypes.Float32)
	vectorArray := array.NewFixedSizeListData(
		array.NewData(vectorListType, 1, []*memory.Buffer{nil},
			[]arrow.ArrayData{vectorFloat32Array.Data()}, 0, 0),
	)
	defer vectorArray.Release()

	fields := []arrow.Field{
		{Name: "id", Type: arrow.BinaryTypes.String, Nullable: false},
		{Name: "theme", Type: arrow.BinaryTypes.String, Nullable: false},
		{Name: "text", Type: arrow.BinaryTypes.String, Nullable: false},
		{Name: "vector", Type: vectorListType, Nullable: false},
	}
	schema := arrow.NewSchema(fields, nil)
	record := array.NewRecord(schema, []arrow.Array{idArray, themeArray, textArray, vectorArray}, 1)
	defer record.Release()

	return LanceTable.AddRecords(ctx, []arrow.Record{record}, nil)
}

// SearchDocument 检索相关的文档片段
func SearchDocument(ctx context.Context, theme string, query string, topK int) ([]string, error) {
	if LanceTable == nil {
		return nil, fmt.Errorf("LanceTable not initialized")
	}

	queryVector, err := getEmbedding(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to embed query: %w", err)
	}

	results, err := LanceTable.Search(queryVector).
		Where(fmt.Sprintf("theme = '%s'", theme)). // 过滤当前幻灯片主题的内容
		Limit(topK).
		Execute(ctx)
	if err != nil {
		return nil, err
	}

	var chunks []string
	for _, res := range results {
		if textVal, ok := res["text"].(string); ok {
			chunks = append(chunks, textVal)
		}
	}

	return chunks, nil
}
