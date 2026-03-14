package render

import (
	"context"
	"github.com/cloudwego/eino/compose"
	"log"
	"ppt-stasher-backend/internal/config"
	"ppt-stasher-backend/internal/workflow/render/subagents"
)

func BuildRenderTeamGraph() *compose.Graph[TeamRenderState, TeamRenderState] {
	g := compose.NewGraph[TeamRenderState, TeamRenderState]()
	modelID := config.GlobalConfig.LLM.CoderModel
	log.Printf("RenderTeam Model initialized with %s", modelID)

	_ = g.AddLambdaNode("script_coder", subagents.NewScriptCoderNode())
	_ = g.AddLambdaNode("ppteval_judge", subagents.NewPPTEvalJudgeNode())

	_ = g.AddEdge(compose.START, "script_coder")
	_ = g.AddEdge("script_coder", "ppteval_judge")
	_ = g.AddEdge("ppteval_judge", compose.END)

	return g
}
