package template

import (
	"context"
	"github.com/cloudwego/eino/compose"
	"log"
	"ppt-stasher-backend/internal/config"
	"ppt-stasher-backend/internal/workflow/template/subagents"
)

func BuildTemplateAnalystGraph() *compose.Graph[TeamTemplateState, TeamTemplateState] {
	g := compose.NewGraph[TeamTemplateState, TeamTemplateState]()
	modelID := config.GlobalConfig.LLM.AnalystModel
	log.Printf("TemplateAnalyst Model initialized with %s", modelID)

	_ = g.AddLambdaNode("cluster_layout", subagents.NewClusterLayoutNode())
	_ = g.AddLambdaNode("schema_extractor", subagents.NewSchemaExtractorNode())
	_ = g.AddLambdaNode("html_renderer", subagents.NewHTMLRendererNode())

	_ = g.AddEdge(compose.START, "cluster_layout")
	_ = g.AddEdge("cluster_layout", "schema_extractor")
	_ = g.AddEdge("schema_extractor", "html_renderer")
	_ = g.AddEdge("html_renderer", compose.END)

	return g
}
