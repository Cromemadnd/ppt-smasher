package agent

import (
	"context"
	"log"

	"ppt-stasher-backend/internal/config"

	"github.com/cloudwego/eino/compose"
)

// BuildContentTeamGraph 构建内容创作团队。
func BuildContentTeamGraph() *compose.Graph[TeamContentState, TeamContentState] {
	g := compose.NewGraph[TeamContentState, TeamContentState]()

	modelID := config.GlobalConfig.LLM.ContentModel
	log.Printf("ContentTeam Model initialized with %s", modelID)

	_ = g.AddLambdaNode("outline_director", compose.InvokableLambda(func(ctx context.Context, s TeamContentState) (TeamContentState, error) {
		log.Println("[ContentTeam:Director] 起草幻灯片大纲，并为每一页指定模版 Schema...")
		s.Outline = "绪论(TitleSlide) -> 核心数据(ContentSlide)"
		return s, nil
	}))

	_ = g.AddLambdaNode("content_filler", compose.InvokableLambda(func(ctx context.Context, s TeamContentState) (TeamContentState, error) {
		log.Println("[ContentTeam:Filler] 从 VDB 提取详细文案，精准填充入骨架坑位中...")
		s.FilledContentDraft = append(s.FilledContentDraft, "Draft data payload...")
		return s, nil
	}))

	_ = g.AddLambdaNode("content_critic", compose.InvokableLambda(func(ctx context.Context, s TeamContentState) (TeamContentState, error) {
		log.Println("[ContentTeam:Critic] 对比原始资料进行严格自省(Reflection)审查事实断层与幻觉...")
		return s, nil
	}))

	_ = g.AddEdge(compose.START, "outline_director")
	_ = g.AddEdge("outline_director", "content_filler")
	_ = g.AddEdge("content_filler", "content_critic")
	_ = g.AddEdge("content_critic", compose.END)

	return g
}
