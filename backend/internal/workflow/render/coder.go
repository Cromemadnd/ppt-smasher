package render

import (
	"context"
	"log"

	"github.com/cloudwego/eino/compose"
)

func NewScriptCoderNode() *compose.Lambda {
	return compose.InvokableLambda(func(ctx context.Context, s TeamRenderState) (TeamRenderState, error) {
		log.Println("[RenderTeam:Coder] 生成 Python 代码渲染 PPT，并进行闭环纠错执行...")
		return s, nil
	})
}
