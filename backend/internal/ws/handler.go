package ws

import (
	"encoding/json"
	"log"
	"ppt-stasher-backend/internal/agent"
)

// HandleClientMessage 处理来自客户端的 WebSocket 消息帧
func HandleClientMessage(c *Client, msg []byte) {
	var m Message
	if err := json.Unmarshal(msg, &m); err != nil {
		log.Printf("json unmarshal err: %v", err)
		return
	}

	switch m.Type {
	case MsgTypeStartTask:
		// 解析 payload
		payloadBytes, _ := json.Marshal(m.Payload)
		var p StartTaskPayload
		if err := json.Unmarshal(payloadBytes, &p); err != nil {
			log.Printf("parse start_task payload error: %v", err)
			return
		}

		// 异步触发 Eino 工作流，直接向当前的 Client 推流
		go agent.RunWorkflow(c, p.Theme)

	case MsgTypeUploadFile:
		// 1. 解析 base64Data
		// 2. 存盘
		// 3. 调大模型生成摘要
		// 4. db存库
		// 5. 将新资料回推或广播
		c.Send <- Message{
			Type: MsgTypeKnowledgeBaseUpdate,
			Payload: map[string]string{
				"fileName": "demo.pdf",
				"desc":     "模型生成的文档摘要",
			},
		}

	case MsgTypeUpdateFileDesc:
		// TODO 更新 LanceDB
	default:
		log.Printf("Unknown message type: %s", m.Type)
	}
}
