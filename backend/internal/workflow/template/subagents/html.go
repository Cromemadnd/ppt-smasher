package subagents

import (
	"context"
	"github.com/cloudwego/eino/compose"
	"log"
	"ppt-stasher-backend/internal/workflow/template"
)

func NewHTMLRendererNode() compose.InvokableLambda[template.TeamTemplateState, template.TeamTemplateState] {
	return compose.InvokableLambda(func(ctx context.Context, s template.TeamTemplateState) (template.TeamTemplateState, error) {
		log.Println("[TemplateAnalyst] 正在渲染为 HTML 视图导航......")
		return s, nil
	})
}
