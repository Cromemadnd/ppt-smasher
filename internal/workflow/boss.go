package workflow

import (
	"context"
	"log"
	"ppt-smasher/internal/config"
	"ppt-smasher/internal/workflow/content"
	"ppt-smasher/internal/workflow/render"
	"ppt-smasher/internal/workflow/research"
	"ppt-smasher/internal/workflow/template"

	"github.com/cloudwego/eino/compose"
)

// BuildBossGraph 构建 Boss 作为最高层级的 Graph 编排，
// 也是基于 ReAct State Orchestrator 来动态分发子任务给不同的团队。
func BuildBossGraph() (compose.Runnable[WorkflowState, WorkflowState], error) {
	g := compose.NewGraph[WorkflowState, WorkflowState]()

	bossModelID := config.GlobalConfig.LLM.BossModel
	log.Printf("Boss Model initialized with %s", bossModelID)

	// 获取各个独立编排子团队
	researchGraph, err := research.BuildResearchTeamGraph().Compile(context.Background())
	if err != nil {
		return nil, err
	}
	templateGraph, err := template.BuildTemplateAnalystGraph().Compile(context.Background())
	if err != nil {
		return nil, err
	}
	contentGraph, err := content.BuildContentTeamGraph().Compile(context.Background())
	if err != nil {
		return nil, err
	}
	renderGraph, err := render.BuildRenderTeamGraph().Compile(context.Background())
	if err != nil {
		return nil, err
	}

	_ = g.AddLambdaNode("boss_reasoning", compose.InvokableLambda(func(ctx context.Context, s WorkflowState) (WorkflowState, error) {
		log.Printf("[Boss] 分析用户需求，主题: '%s', 然后决定调度哪些下层 Agent.", s.Theme)
		return s, nil
	}))

	_ = g.AddLambdaNode("call_tool_research", compose.InvokableLambda(func(ctx context.Context, s WorkflowState) (WorkflowState, error) {
		log.Println("[Boss -> Tool Call] 召唤 Research Team...")
		rs, _ := researchGraph.Invoke(ctx, research.TeamResearchState{Theme: s.Theme, GivenDocuments: s.GivenDocuments})
		s.KnowledgeReady = rs.VDBStatus
		return s, nil
	}))

	_ = g.AddLambdaNode("call_tool_template", compose.InvokableLambda(func(ctx context.Context, s WorkflowState) (WorkflowState, error) {
		log.Println("[Boss -> Tool Call] 召唤 Template Analyst...")
		ts, _ := templateGraph.Invoke(ctx, template.TeamTemplateState{ReferencePPT: s.ReferencePPT})
		s.LayoutSchemas = ts.Schemas
		return s, nil
	}))

	_ = g.AddLambdaNode("call_tool_content", compose.InvokableLambda(func(ctx context.Context, s WorkflowState) (WorkflowState, error) {
		log.Println("[Boss -> Tool Call] 召唤 Content Team 开始共创文案大纲与版式分配...")
		cs, _ := contentGraph.Invoke(ctx, content.TeamContentState{
			Theme:            s.Theme,
			VDBStatus:        s.KnowledgeReady,
			AvailableLayouts: s.LayoutSchemas,
		})
		s.ContentDrafts = cs.FilledContentDraft
		s.Outline = cs.Outline
		return s, nil
	}))

	_ = g.AddLambdaNode("call_tool_render", compose.InvokableLambda(func(ctx context.Context, s WorkflowState) (WorkflowState, error) {
		log.Println("[Boss -> Tool Call] 召唤 Render Team 编写 Python 代码渲染出 PPTX...")
		rs, _ := renderGraph.Invoke(ctx, render.TeamRenderState{ContentDrafts: s.ContentDrafts})
		s.PPTXFiles = rs.RenderResults
		return s, nil
	}))

	_ = g.AddEdge(compose.START, "boss_reasoning")
	// 模板分析与资料研究可以并行，而内容生成需要前两者的结果
	_ = g.AddEdge("boss_reasoning", "call_tool_research")
	_ = g.AddEdge("boss_reasoning", "call_tool_template")
	_ = g.AddEdge("call_tool_research", "call_tool_content")
	_ = g.AddEdge("call_tool_template", "call_tool_content")

	_ = g.AddEdge("call_tool_content", "call_tool_render")
	_ = g.AddEdge("call_tool_render", compose.END)

	return g.Compile(context.Background())
}
