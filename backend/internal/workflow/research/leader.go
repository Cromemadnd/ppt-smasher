package research

import (
	"log"
	"ppt-stasher-backend/internal/config"

	"github.com/cloudwego/eino/compose"
)

func BuildResearchTeamGraph() *compose.Graph[TeamResearchState, TeamResearchState] {
	g := compose.NewGraph[TeamResearchState, TeamResearchState]()
	modelID := config.GlobalConfig.LLM.ResearcherModel
	log.Printf("ResearchTeam Model initialized with %s", modelID)

	_ = g.AddLambdaNode("search_document", NewSearchDocumentNode())
	_ = g.AddLambdaNode("search_image", NewSearchImageNode())
	_ = g.AddLambdaNode("search_analytics", NewSearchAnalyticsNode())

	_ = g.AddEdge(compose.START, "search_document")
	_ = g.AddEdge("search_document", "search_image")
	_ = g.AddEdge("search_image", "search_analytics")
	_ = g.AddEdge("search_analytics", compose.END)

	return g
}
