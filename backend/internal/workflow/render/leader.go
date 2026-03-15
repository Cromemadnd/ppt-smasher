package render

import (
	"log"
	"ppt-stasher-backend/internal/config"

	"github.com/cloudwego/eino/compose"
)

func BuildRenderTeamGraph() *compose.Graph[TeamRenderState, TeamRenderState] {
	g := compose.NewGraph[TeamRenderState, TeamRenderState]()
	modelID := config.GlobalConfig.LLM.VisualModel
	log.Printf("RenderTeam Model initialized with %s", modelID)

	_ = g.AddLambdaNode("script_coder", NewScriptCoderNode())
	_ = g.AddLambdaNode("ppteval_judge", NewPPTEvalJudgeNode())

	_ = g.AddEdge(compose.START, "script_coder")
	_ = g.AddEdge("script_coder", "ppteval_judge")
	_ = g.AddEdge("ppteval_judge", compose.END)

	return g
}
