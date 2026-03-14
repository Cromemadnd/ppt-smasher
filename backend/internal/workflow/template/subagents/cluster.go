package subagents

import (
	"context"
	"github.com/cloudwego/eino/compose"
	"log"
	"ppt-stasher-backend/internal/workflow/template"
)

func NewClusterLayoutNode() compose.InvokableLambda[template.TeamTemplateState, template.TeamTemplateState] {
	return compose.InvokableLambda(func(ctx context.Context, s template.TeamTemplateState) (template.TeamTemplateState, error) {
		log.Println("[TemplateAnalyst] 正在切分聚类，分析幻灯片功能类型...")
		return s, nil
	})
}
