package agent

import (
	"context"
	"log"

	"ppt-stasher-backend/internal/config"

	"github.com/cloudwego/eino/compose"
)

// BuildRenderTeamGraph 视觉设计团队 (包含 Script Coder 和 PPTEval Judge)
func BuildRenderTeamGraph() *compose.Graph[TeamRenderState, TeamRenderState] {
	g := compose.NewGraph[TeamRenderState, TeamRenderState]()

	modelID := config.GlobalConfig.LLM.VisualModel
	log.Printf("RenderTeam Model initialized with %s", modelID)

	_ = g.AddLambdaNode("script_coder", compose.InvokableLambda(func(ctx context.Context, s TeamRenderState) (TeamRenderState, error) {
		log.Println("[RenderTeam:ScriptCoder] 编写代码操作本地沙盒调用 PPTX...")
		log.Println("[REPL] Sandbox Execution (Mock): Editing via Python Agent repl.")
		// 实际上会调用 /python/server.py 里的 sandbox 
		s.RenderResults = append(s.RenderResults, "generated_output.pptx")
		return s, nil
	}))

	_ = g.AddLambdaNode("eval_judge", compose.InvokableLambda(func(ctx context.Context, s TeamRenderState) (TeamRenderState, error) {
		log.Println("[RenderTeam:PPTEvalJudge] 利用规则打分，验收 Content, Design, Coherence 三大维度...")
		return s, nil
	}))

	_ = g.AddEdge(compose.START, "script_coder")
	_ = g.AddEdge("script_coder", "eval_judge")
	_ = g.AddEdge("eval_judge", compose.END)

	return g
}
