package main

import (
	"context"
	"log"
	"net/http"
	"ppt-smasher/internal/config"
	"ppt-smasher/internal/db"
	"ppt-smasher/internal/llm"
	"ppt-smasher/internal/workflow"
	"ppt-smasher/internal/ws"
)

func main() {
	// 0. 初始化配置
	config.InitConfig([]string{"."})

	// 初始化 Postgres + pgvector 向量库
	db.InitVectorDB(context.Background())

	// 初始化 LLM 模型
	llm.InitChatModels(context.Background())

	// 1.1 初始化 Agent 编排
	if err := workflow.InitWorkflow(); err != nil {
		log.Fatalf("InitWorkflow error: %v", err)
	}

	// 2. 初始化 Websocket 消息枢纽
	hub := ws.NewHub()
	go hub.Run()

	// 3. 注册 WebSocket 路由
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		ws.ServeWs(hub, w, r)
	})

	// 4. 启动后端服务器
	addr := ":8080"
	log.Printf("Backend server staring on %s...", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Server ListenAndServe error: %v", err)
	}
}
