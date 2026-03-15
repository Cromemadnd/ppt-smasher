package render

import (
	"context"
	"log"

	"github.com/cloudwego/eino/compose"
)

func NewPPTEvalJudgeNode() *compose.Lambda {
	return compose.InvokableLambda(func(ctx context.Context, s TeamRenderState) (TeamRenderState, error) {
		log.Println("[RenderTeam:Coder] 终审 Judge (Content/Design/Coherence)打分......")
		return s, nil
	})
}
