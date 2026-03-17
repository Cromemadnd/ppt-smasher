package ws

type Message struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

// 客户端 -> 服务端
const (
	MsgTypeStartTask      = "START_TASK"
	MsgTypeUploadFile     = "UPLOAD_FILE"
	MsgTypeUpdateFileDesc = "UPDATE_FILE_DESC"
)

// 服务端 -> 客户端
const (
	MsgTypeKnowledgeBaseUpdate = "KNOWLEDGE_BASE_UPDATE"
	MsgTypeNodeActive          = "NODE_ACTIVE"
	MsgTypeEdgeActive          = "EDGE_ACTIVE"
	MsgTypeAgentThoughtStream  = "AGENT_THOUGHT_STREAM"
	MsgTypeStepCompleted       = "STEP_COMPLETED"
	MsgTypeTaskSuccess         = "TASK_SUCCESS"
	MsgTypeTaskError           = "TASK_ERROR"
)

type StartTaskPayload struct {
	Theme string `json:"theme"`
}

type UploadFilePayload struct {
	FileName   string `json:"fileName"`
	FileType   string `json:"fileType"`
	Base64Data string `json:"base64Data"`
}

type StreamPayload struct {
	NodeID string `json:"nodeId"`
	Chunk  string `json:"chunk"`
	Status string `json:"status"` // pending, done
}
