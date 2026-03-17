package db

import (
	"context"
	"fmt"
	"log"

	"ppt-smasher/internal/config"

	"github.com/cloudwego/eino/components/embedding"
	einoindexer "github.com/cloudwego/eino/components/indexer"
	einoretriever "github.com/cloudwego/eino/components/retriever"
	"github.com/cloudwego/eino/schema"
	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
)

type MilvusVectorDB struct {
	client     client.Client
	collection string
	dimension  int
	embedder   embedding.Embedder
}

func (m *MilvusVectorDB) Close() error {
	return m.client.Close()
}

func NewMilvusVectorDB(ctx context.Context, cfg *config.VDBConfig, embedder embedding.Embedder, collectionName string, dimension int) *MilvusVectorDB {
	// 初始连接使用配置信息
	clientCfg := client.Config{
		Address:  fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Username: cfg.User,
		Password: cfg.Password,
	}

	c, err := client.NewClient(ctx, clientCfg)
	if err != nil {
		log.Fatalf("failed to connect to Milvus: %v", err)
	}

	// 如果指定了 DBName 且不是默认数据库，检查是否存在并尝试创建
	if cfg.DBName != "" && cfg.DBName != "default" {
		dbs, err := c.ListDatabases(ctx)
		if err != nil {
			log.Printf("warning: failed to list databases: %v", err)
		} else {
			exists := false
			for _, db := range dbs {
				if db.Name == cfg.DBName {
					exists = true
					break
				}
			}
			if !exists {
				log.Printf("database %s not found, creating...", cfg.DBName)
				if err := c.CreateDatabase(ctx, cfg.DBName); err != nil {
					log.Printf("warning: failed to create database %s: %v (it might already exist or permission denied)", cfg.DBName, err)
				}
			}
		}

		// 重新连接或切换到目标数据库
		c.Close()
		clientCfg.DBName = cfg.DBName
		c, err = client.NewClient(ctx, clientCfg)
		if err != nil {
			log.Fatalf("failed to connect to Milvus database %s: %v", cfg.DBName, err)
		}
	}

	mvdb := &MilvusVectorDB{
		client:     c,
		collection: collectionName,
		dimension:  dimension,
		embedder:   embedder,
	}

	// 检查 collection 是否存在，不存在则创建
	has, err := c.HasCollection(ctx, collectionName)
	if err != nil {
		log.Fatalf("failed to check collection: %v", err)
	}

	if !has {
		schema := &entity.Schema{
			CollectionName: collectionName,
			Description:    "ppt-smasher vector storage",
			Fields: []*entity.Field{
				{
					Name:        "id",
					DataType:    entity.FieldTypeInt64,
					PrimaryKey:  true,
					AutoID:      true,
					Description: "primary key",
				},
				{
					Name:        "vector",
					DataType:    entity.FieldTypeFloatVector,
					TypeParams:  map[string]string{"dim": fmt.Sprintf("%d", dimension)},
					Description: "vector",
				},
				{
					Name:        "content",
					DataType:    entity.FieldTypeVarChar,
					TypeParams:  map[string]string{"max_length": "65535"},
					Description: "content",
				},
				{
					Name:        "metadata",
					DataType:    entity.FieldTypeJSON,
					Description: "metadata",
				},
			},
		}
		err = c.CreateCollection(ctx, schema, entity.DefaultShardNumber)
		if err != nil {
			log.Fatalf("failed to create collection: %v", err)
		}

		// 创建索引
		idx, err := entity.NewIndexIvfFlat(entity.L2, 1024)
		if err != nil {
			log.Fatalf("failed to create index entity: %v", err)
		}
		err = c.CreateIndex(ctx, collectionName, "vector", idx, false)
		if err != nil {
			log.Fatalf("failed to create index: %v", err)
		}
	}

	// 加载 collection
	err = c.LoadCollection(ctx, collectionName, false)
	if err != nil {
		log.Fatalf("failed to load collection: %v", err)
	}

	return mvdb
}

// Index 实现 einoindexer.Indexer 接口
func (m *MilvusVectorDB) Index(ctx context.Context, docs []*schema.Document, opts ...einoindexer.Option) ([]string, error) {
	texts := make([]string, 0, len(docs))
	for _, doc := range docs {
		texts = append(texts, doc.Content)
	}

	vectors, err := m.embedder.EmbedStrings(ctx, texts)
	if err != nil {
		return nil, fmt.Errorf("failed to embed strings: %v", err)
	}

	contentData := make([]string, 0, len(docs))
	metadataData := make([][]byte, 0, len(docs))
	vectorData := make([][]float32, 0, len(docs))

	for i, doc := range docs {
		contentData = append(contentData, doc.Content)
		// Convert []float64 to []float32 for Milvus
		v64 := vectors[i]
		v32 := make([]float32, len(v64))
		for j, v := range v64 {
			v32[j] = float32(v)
		}
		vectorData = append(vectorData, v32)
		// 简单处理 metadata 为 JSON，Milvus 支持 JSON 类型
		// 这里暂不深入实现复杂的 metadata 映射
		metadataData = append(metadataData, []byte("{}"))
	}

	contentCol := entity.NewColumnVarChar("content", contentData)
	vectorCol := entity.NewColumnFloatVector("vector", m.dimension, vectorData)
	metadataCol := entity.NewColumnJSONBytes("metadata", metadataData)

	_, err = m.client.Insert(ctx, m.collection, "", contentCol, vectorCol, metadataCol)
	if err != nil {
		return nil, fmt.Errorf("failed to insert data into milvus: %v", err)
	}

	// Milvus Insert 不直接返回 string ID (如果是 AutoID)
	// 为了简化，返回 dummy IDs 或者在实际应用中根据需求调整
	ids := make([]string, len(docs))
	for i := range docs {
		ids[i] = fmt.Sprintf("%d", i)
	}

	return ids, nil
}

// Retrieve 实现 einoretriever.Retriever 接口
func (m *MilvusVectorDB) Retrieve(ctx context.Context, query string, opts ...einoretriever.Option) ([]*schema.Document, error) {
	vectors, err := m.embedder.EmbedStrings(ctx, []string{query})
	if err != nil {
		return nil, fmt.Errorf("failed to embed query: %v", err)
	}

	// Convert []float64 to []float32 for Milvus
	v64 := vectors[0]
	v32 := make([]float32, len(v64))
	for i, v := range v64 {
		v32[i] = float32(v)
	}

	// Use client.NewSearchRequest to avoid manual field issues
	searchParam, _ := entity.NewIndexIvfFlatSearchParam(10)
	searchResult, err := m.client.Search(ctx, m.collection, []string{}, "", []string{"content", "metadata"}, []entity.Vector{entity.FloatVector(v32)}, "vector", entity.L2, 10, searchParam)
	if err != nil {
		return nil, fmt.Errorf("failed to search in milvus: %v", err)
	}

	var docs []*schema.Document
	for _, res := range searchResult {
		contentCol := res.Fields.GetColumn("content")
		for i := 0; i < contentCol.Len(); i++ {
			val, _ := contentCol.Get(i)
			docs = append(docs, &schema.Document{
				Content: val.(string),
			})
		}
	}

	return docs, nil
}

func (m *MilvusVectorDB) AddDocumentChunk(ctx context.Context, chunks []string) error {
	docs := make([]*schema.Document, len(chunks))
	for i, chunk := range chunks {
		docs[i] = &schema.Document{
			Content: chunk,
		}
	}
	_, err := m.Index(ctx, docs)
	return err
}

func (m *MilvusVectorDB) SearchDocument(ctx context.Context, query string) ([]string, error) {
	docs, err := m.Retrieve(ctx, query)
	if err != nil {
		return nil, err
	}

	results := make([]string, len(docs))
	for i, doc := range docs {
		results[i] = doc.Content
	}
	return results, nil
}
