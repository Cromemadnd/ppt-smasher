package db

import (
	"context"
	"log"

	"github.com/apache/arrow/go/v17/arrow"
	"github.com/lancedb/lancedb-go/pkg/lancedb"
)

var LanceDB *lancedb.Connection
var LanceTable *lancedb.Table

// InitLanceDB 初始化 LanceDB 向量数据库连接
func InitLanceDB() {
	db, err := lancedb.Connect(context.Background(), "data/lancedb_store", nil)
	if err != nil {
		log.Fatalf("failed opening connection to lancedb: %v", err)
	}

	LanceDB = db
	log.Println("LanceDB memory initialized successfully")

	// 初始化一个 RAG 专用的文档索引 Table
	initRAGTable()
}

func initRAGTable() {
	ctx := context.Background()

	// 简单的 schema：id (string), text (string), vector (float32[384])
	fields := []arrow.Field{
		{Name: "id", Type: arrow.BinaryTypes.String, Nullable: false},
		{Name: "text", Type: arrow.BinaryTypes.String, Nullable: false},
		{Name: "vector", Type: arrow.FixedSizeListOf(384, arrow.PrimitiveTypes.Float32), Nullable: false},
	}
	schema := arrow.NewSchema(fields, nil)
	ldbSchema, _ := lancedb.NewSchema(schema)

	tableName := "document_chunks"
	tableNames, _ := LanceDB.TableNames(ctx, nil)
	
	// 如果不存在，则创建表
	exists := false
	for _, name := range tableNames {
		if name == tableName {
			exists = true
			break
		}
	}

	var table *lancedb.Table
	var err error
	if exists {
		table, err = LanceDB.OpenTable(ctx, tableName)
		if err != nil {
			log.Fatalf("failed to open lancedb table: %v", err)
		}
	} else {
		table, err = LanceDB.CreateTable(ctx, tableName, ldbSchema)
		if err != nil {
			log.Fatalf("failed to create lancedb table: %v", err)
		}
	}
	
	LanceTable = table
}
