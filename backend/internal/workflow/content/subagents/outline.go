package subagents

import (
	"context"
	"github.com/cloudwego/eino/compose"
	"log"
	"ppt-stasher-backend/internal/workflow/content"
)

func NewOutlineDirectorNode() compose.InvokableLambda[content.TeamContentState, content.TeamContentState] {
	return compose.InvokableLambda(func(ctx context.Context, s content.TeamContentState) (content.TeamContentState, error) {
		log.Println("[ContentTeam:Director] 起草幻灯片大纲，选择模板Schema...")
		s.Outline = "绪论(TitleSlide) -> 核心数据(ContentSlide)"
		return s, nil
	})
}
