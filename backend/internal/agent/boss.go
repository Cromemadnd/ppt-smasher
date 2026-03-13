package agent

import (
	"context"
	"log"

	"ppt-stasher-backend/internal/config"

	"github.com/cloudwego/eino/compose"
)

// BuildBossGraph 构建 Boss 作为最高层级的 Graph 编排
// 在这里 Boss 会将下位 Agent (Researcher, ContentLeader, VisualLeader) 作为 Tool 或下游节点调用。
func BuildBossGraph() (compose.Runnable[WorkflowState, WorkflowState], error) {
	g := compose.NewGraph[WorkflowState, WorkflowState]()

	// 1. 初始化 Boss 节点以及它手下的下位 Agent (也是 Graph)
	bossModelID := config.GlobalConfig.LLM.BossModel
	log.Printf("Boss Model initialized with %s", bossModelID)

	// 获取下位 Agent 的编排 Runnable
	researcherGraph := BuildResearcherGraph()
	contentGraph := BuildContentLeaderGraph()
	visualGraph := BuildVisualGraph()

	// 2. 将下位 Graph 作为节点注册（上位 Agent 调用它们。这里我们用节点模拟工具调用流程）
	// 在真实的纯 Agent 场景中，可以使用 BindTools 把 Runnable 封装为 Tool 给 Boss 的大模型调用，
	// 此处为示范：Boss 作路由/拆解，依次唤起各个子 Graph 处理具体逻辑。
	_ = g.AddLambdaNode("boss_node", compose.InvokableLambda(func(ctx context.Context, s WorkflowState) (WorkflowState, error) {
		log.Printf("[Boss] 接收到任务，开始规划拆解，主题: %s", s.Theme)
		// Boss 处理自身逻辑...
		return s, nil
	}))

	_ = g.AddGraphNode("researcher_node", researcherGraph)
	_ = g.AddGraphNode("content_node", contentGraph)
	_ = g.AddGraphNode("visual_node", visualGraph)

	// 3. 编排连线 (Boss -> Researcher -> Content -> Visual)
	_ = g.AddEdge(compose.START, "boss_node")
	_ = g.AddEdge("boss_node", "researcher_node")
	_ = g.AddEdge("researcher_node", "content_node")
	_ = g.AddEdge("content_node", "visual_node")
	_ = g.AddEdge("visual_node", compose.END)

	return g.Compile(context.Background())
}
