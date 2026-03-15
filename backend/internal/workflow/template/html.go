package template

import (
	"context"
	"log"

	"github.com/cloudwego/eino/compose"
)

func NewHTMLRendererNode() *compose.Lambda {
	return compose.InvokableLambda(func(ctx context.Context, s TeamTemplateState) (TeamTemplateState, error) {
		log.Println("[TemplateAnalyst] 正在渲染为 HTML 视图导航......")
		return s, nil
	})
}
