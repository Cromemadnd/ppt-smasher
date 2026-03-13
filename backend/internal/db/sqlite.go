package db

import (
	"context"
	"log"

	"ppt-stasher-backend/ent"

	_ "github.com/mattn/go-sqlite3" // sqlite3 驱动
)

var Client *ent.Client

func InitDB() {
	// 开启外键支持：?_fk=1
	client, err := ent.Open("sqlite3", "file:ppt_stasher.db?cache=shared&_fk=1")
	if err != nil {
		log.Fatalf("failed opening connection to sqlite: %v", err)
	}

	// 运行 schema 自动迁移
	if err := client.Schema.Create(context.Background()); err != nil {
		log.Fatalf("failed creating schema resources: %v", err)
	}

	Client = client
	log.Println("SQLite database initialized successfully")
}
