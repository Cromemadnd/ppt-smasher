package content

import (
	"log"
	"ppt-stasher-backend/internal/config"

	"github.com/cloudwego/eino/compose"
)

func BuildContentTeamGraph() *compose.Graph[TeamContentState, TeamContentState] {
	g := compose.NewGraph[TeamContentState, TeamContentState]()
	modelID := config.GlobalConfig.LLM.ContentModel
	log.Printf("ContentTeam Model initialized with %s", modelID)

	_ = g.AddLambdaNode("outline_director", NewOutlineDirectorNode())
	_ = g.AddLambdaNode("content_filler", NewContentFillerNode())
	_ = g.AddLambdaNode("content_critic", NewContentCriticNode())

	_ = g.AddEdge(compose.START, "outline_director")
	_ = g.AddEdge("outline_director", "content_filler")
	_ = g.AddEdge("content_filler", "content_critic")
	_ = g.AddEdge("content_critic", compose.END)

	return g
}
