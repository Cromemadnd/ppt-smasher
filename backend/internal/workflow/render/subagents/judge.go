package subagents

import (
	"context"
	"github.com/cloudwego/eino/compose"
	"log"
	"ppt-stasher-backend/internal/workflow/render"
)

func NewPPTEvalJudgeNode() compose.InvokableLambda[render.TeamRenderState, render.TeamRenderState] {
	return compose.InvokableLambda(func(ctx context.Context, s render.TeamRenderState) (render.TeamRenderState, error) {
		log.Println("[RenderTeam:Coder] 终审 Judge (Content/Design/Coherence)打分......")
		return s, nil
	})
}
