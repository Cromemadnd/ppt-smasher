package agent

import (
	"context"
	"log"

	"ppt-stasher-backend/internal/config"

	"github.com/cloudwego/eino/compose"
)

// BuildContentLeaderGraph 构建 ContentLeader 节点编排。
func BuildContentLeaderGraph() *compose.Graph[WorkflowState, WorkflowState] {
	g := compose.NewGraph[WorkflowState, WorkflowState]()

	modelID := config.GlobalConfig.LLM.ContentModel
	log.Printf("ContentLeader Model initialized with %s", modelID)

	_ = g.AddLambdaNode("content_draft", compose.InvokableLambda(func(ctx context.Context, s WorkflowState) (WorkflowState, error) {
		log.Println("[ContentLeader] 正在提炼大纲与内容...")
		s.Outline = "1. 绪论 2. 核心内容 3. 结论"
		return s, nil
	}))

	_ = g.AddEdge(compose.START, "content_draft")
	_ = g.AddEdge("content_draft", compose.END)

	return g
}
