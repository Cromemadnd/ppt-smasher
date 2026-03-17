package render

import (
	"context"
	"log"

	"github.com/cloudwego/eino/compose"
)

func NewPPTEvalJudgeNode() *compose.Lambda {
	return compose.InvokableLambda(func(ctx context.Context, s TeamRenderState) (TeamRenderState, error) {
		log.Println("[RenderTeam:Judge] 终审 Judge: 检查生成的 PPT 是否成功。")

		if len(s.RenderResults) > 0 {
			log.Printf("[RenderTeam:Judge] 发现生成的 PPTX: %v，打分通过。", s.RenderResults)
		} else {
			log.Println("[RenderTeam:Judge] 警告: 未找到渲染结果。")
		}

		return s, nil
	})
}
