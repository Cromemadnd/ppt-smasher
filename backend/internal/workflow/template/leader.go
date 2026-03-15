package template

import (
	"log"
	"ppt-stasher-backend/internal/config"

	"github.com/cloudwego/eino/compose"
)

func BuildTemplateAnalystGraph() *compose.Graph[TeamTemplateState, TeamTemplateState] {
	g := compose.NewGraph[TeamTemplateState, TeamTemplateState]()
	modelID := config.GlobalConfig.LLM.BossModel
	log.Printf("TemplateAnalyst Model initialized with %s", modelID)

	_ = g.AddLambdaNode("cluster_layout", NewClusterLayoutNode())
	_ = g.AddLambdaNode("schema_extractor", NewSchemaExtractorNode())
	_ = g.AddLambdaNode("html_renderer", NewHTMLRendererNode())

	_ = g.AddEdge(compose.START, "cluster_layout")
	_ = g.AddEdge("cluster_layout", "schema_extractor")
	_ = g.AddEdge("schema_extractor", "html_renderer")
	_ = g.AddEdge("html_renderer", compose.END)

	return g
}
