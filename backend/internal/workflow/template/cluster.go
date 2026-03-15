package template

import (
	"context"
	"log"

	"github.com/cloudwego/eino/compose"
)

func NewClusterLayoutNode() *compose.Lambda {
	return compose.InvokableLambda(func(ctx context.Context, s TeamTemplateState) (TeamTemplateState, error) {
		// PPTAgent 原始逻辑中可能有复杂的聚类算法。
		// Eino 重构版本中：当前按页提取的 Schema 已足够在 Content Director 分配使用，
		// 各个版本通过 LLM 抽取出的 LayoutName 足以实现语义分类，故此处直接透传跳过深度聚类。
		log.Println("[TemplateAnalyst:Cluster] 版式切分与语义化(LayoutName)聚类已在Schema节点中基本完成，执行透传...")
		return s, nil
	})
}
