package template

import (
	"context"
	"log"

	"github.com/cloudwego/eino/compose"
)

func NewClusterLayoutNode() *compose.Lambda {
	return compose.InvokableLambda(func(ctx context.Context, s TeamTemplateState) (TeamTemplateState, error) {
		log.Println("[TemplateAnalyst] 正在切分聚类，分析幻灯片功能类型...")
		return s, nil
	})
}
