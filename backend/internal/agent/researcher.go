package agent

import (
	"context"
	"log"

	"ppt-stasher-backend/internal/config"

	"github.com/cloudwego/eino/compose"
)

// BuildResearcherGraph 构建 Researcher 节点。
// Researcher 本身也是一个 Graph，内部可以包含意图识别、搜索引擎、检索知识库等多个子步骤的编排。
func BuildResearcherGraph() *compose.Graph[WorkflowState, WorkflowState] {
	g := compose.NewGraph[WorkflowState, WorkflowState]()

	modelID := config.GlobalConfig.LLM.ResearcherModel
	log.Printf("Researcher Model initialized with %s", modelID)

	// 添加检索逻辑节点
	_ = g.AddLambdaNode("researcher_search", compose.InvokableLambda(func(ctx context.Context, s WorkflowState) (WorkflowState, error) {
		log.Println("[Researcher] 正在搜索文献和资料...")
		s.Researched = append(s.Researched, "资料A: "+s.Theme+" 的背景...")
		return s, nil
	}))

	_ = g.AddEdge(compose.START, "researcher_search")
	_ = g.AddEdge("researcher_search", compose.END)

	return g
}
