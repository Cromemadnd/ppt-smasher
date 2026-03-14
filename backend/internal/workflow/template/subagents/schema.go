package subagents

import (
	"context"
	"github.com/cloudwego/eino/compose"
	"log"
	"ppt-stasher-backend/internal/workflow/template"
)

func NewSchemaExtractorNode() compose.InvokableLambda[template.TeamTemplateState, template.TeamTemplateState] {
	return compose.InvokableLambda(func(ctx context.Context, s template.TeamTemplateState) (template.TeamTemplateState, error) {
		log.Println("[TemplateAnalyst] 正在提取结构骨架(Schema)......")
		return s, nil
	})
}
