package agent

import (
	"context"
	"log"

	"github.com/cloudwego/eino/compose"
)

// BuildTemplateAnalystGraph 构建 Template Analyst 模板分析专家编排。
func BuildTemplateAnalystGraph() *compose.Graph[TeamTemplateState, TeamTemplateState] {
	g := compose.NewGraph[TeamTemplateState, TeamTemplateState]()

	_ = g.AddLambdaNode("slice_and_cluster", compose.InvokableLambda(func(ctx context.Context, s TeamTemplateState) (TeamTemplateState, error) {
		log.Println("[TemplateAnalyst] 正在将幻灯片功能类型进行分析聚合...")
		return s, nil
	}))

	_ = g.AddLambdaNode("extract_schema", compose.InvokableLambda(func(ctx context.Context, s TeamTemplateState) (TeamTemplateState, error) {
		log.Println("[TemplateAnalyst] 正在提取页面数据的排版骨架 (JSON Schema)...")
		s.Schemas = append(s.Schemas, "Schema:TitleSlide, Schema:ContentSlide")
		return s, nil
	}))

	_ = g.AddLambdaNode("render_html", compose.InvokableLambda(func(ctx context.Context, s TeamTemplateState) (TeamTemplateState, error) {
		log.Println("[TemplateAnalyst] 正在生成 HTML 视图映射...")
		return s, nil
	}))

	_ = g.AddEdge(compose.START, "slice_and_cluster")
	_ = g.AddEdge("slice_and_cluster", "extract_schema")
	_ = g.AddEdge("extract_schema", "render_html")
	_ = g.AddEdge("render_html", compose.END)

	return g
}
