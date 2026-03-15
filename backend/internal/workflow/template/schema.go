package template

import (
	"context"
	"log"

	"github.com/cloudwego/eino/compose"
)

func NewSchemaExtractorNode() *compose.Lambda {
	return compose.InvokableLambda(func(ctx context.Context, s TeamTemplateState) (TeamTemplateState, error) {
		log.Println("[TemplateAnalyst] 正在提取结构骨架(Schema)......")
		return s, nil
	})
}
