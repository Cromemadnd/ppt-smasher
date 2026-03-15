package main

import (
	"log"
	"net/http"
	"ppt-stasher-backend/internal/config"
	"ppt-stasher-backend/internal/db"
	"ppt-stasher-backend/internal/workflow"
	"ppt-stasher-backend/internal/ws"
)

func main() {
	// 0. 初始化配置
	config.InitConfig()

	// 初始化 LanceDB 向量库
	db.InitLanceDB()

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
