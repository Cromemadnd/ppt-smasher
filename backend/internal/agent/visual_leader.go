package agent

import (
	"context"
	"log"

	"ppt-stasher-backend/internal/config"

	"github.com/cloudwego/eino/compose"
)

// BuildVisualGraph 构建 Visual 节点编排。
func BuildVisualGraph() *compose.Graph[WorkflowState, WorkflowState] {
	g := compose.NewGraph[WorkflowState, WorkflowState]()

	modelID := config.GlobalConfig.LLM.VisualModel
	log.Printf("Visual Model initialized with %s", modelID)

	_ = g.AddLambdaNode("visual_gen", compose.InvokableLambda(func(ctx context.Context, s WorkflowState) (WorkflowState, error) {
		log.Println("[VisualLeader] 根据大纲生成视觉配图与排版排印...")
		s.VisualLinks = append(s.VisualLinks, "https://mock.com/slide1.png")
		return s, nil
	}))

	_ = g.AddEdge(compose.START, "visual_gen")
	_ = g.AddEdge("visual_gen", compose.END)

	return g
}
