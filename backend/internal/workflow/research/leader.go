package agent

import (
	"context"
	"log"

	"ppt-stasher-backend/internal/config"

	"github.com/cloudwego/eino/compose"
)

// BuildResearchTeamGraph 构建 Research Team 节点编排。
func BuildResearchTeamGraph() *compose.Graph[TeamResearchState, TeamResearchState] {
	g := compose.NewGraph[TeamResearchState, TeamResearchState]()

	modelID := config.GlobalConfig.LLM.ResearcherModel
	log.Printf("ResearchTeam Model initialized with %s", modelID)

	// 搜索和分析节点可以并行或串行
	_ = g.AddLambdaNode("search_document", compose.InvokableLambda(func(ctx context.Context, s TeamResearchState) (TeamResearchState, error) {
		log.Println("[ResearchTeam] 正在检索文本资料...")
		return s, nil
	}))

	_ = g.AddLambdaNode("search_image", compose.InvokableLambda(func(ctx context.Context, s TeamResearchState) (TeamResearchState, error) {
		log.Println("[ResearchTeam] 正在搜集配图与视觉素材...")
		return s, nil
	}))

	_ = g.AddLambdaNode("search_analytics", compose.InvokableLambda(func(ctx context.Context, s TeamResearchState) (TeamResearchState, error) {
		log.Println("[ResearchTeam] 正在抓取数据图表...")
		// 合并所有结果入VDB
		s.VDBStatus = true
		return s, nil
	}))

	// Graph 连线：这里简单做串联模拟，实际中可以结合分支执行 (Branch) 和归并 (Merge)
	_ = g.AddEdge(compose.START, "search_document")
	_ = g.AddEdge("search_document", "search_image")
	_ = g.AddEdge("search_image", "search_analytics")
	_ = g.AddEdge("search_analytics", compose.END)

	return g
}
