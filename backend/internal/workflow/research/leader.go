package research

import (
	"context"
	"github.com/cloudwego/eino/compose"
	"log"
	"ppt-stasher-backend/internal/config"
	"ppt-stasher-backend/internal/workflow/research/subagents"
)

func BuildResearchTeamGraph() *compose.Graph[TeamResearchState, TeamResearchState] {
	g := compose.NewGraph[TeamResearchState, TeamResearchState]()
	modelID := config.GlobalConfig.LLM.ResearcherModel
	log.Printf("ResearchTeam Model initialized with %s", modelID)

	_ = g.AddLambdaNode("search_document", subagents.NewSearchDocumentNode())
	_ = g.AddLambdaNode("search_image", subagents.NewSearchImageNode())
	_ = g.AddLambdaNode("search_analytics", subagents.NewSearchAnalyticsNode())

	_ = g.AddEdge(compose.START, "search_document")
	_ = g.AddEdge("search_document", "search_image")
	_ = g.AddEdge("search_image", "search_analytics")
	_ = g.AddEdge("search_analytics", compose.END)

	return g
}
