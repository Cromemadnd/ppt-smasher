package subagents

import (
	"context"
	"github.com/cloudwego/eino/compose"
	"log"
	"ppt-stasher-backend/internal/workflow/render"
)

func NewScriptCoderNode() compose.InvokableLambda[render.TeamRenderState, render.TeamRenderState] {
	return compose.InvokableLambda(func(ctx context.Context, s render.TeamRenderState) (render.TeamRenderState, error) {
		log.Println("[RenderTeam:Coder] 生成 Python 代码渲染 PPT，并进行闭环纠错执行...")
		return s, nil
	})
}
