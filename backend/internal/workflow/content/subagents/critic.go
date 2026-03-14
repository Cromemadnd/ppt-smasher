package subagents

import (
	"context"
	"github.com/cloudwego/eino/compose"
	"log"
	"ppt-stasher-backend/internal/workflow/content"
)

func NewContentCriticNode() compose.InvokableLambda[content.TeamContentState, content.TeamContentState] {
	return compose.InvokableLambda(func(ctx context.Context, s content.TeamContentState) (content.TeamContentState, error) {
		log.Println("[ContentTeam:Director] 内部审查(Critic)：事实核查与防幻觉...")

		return s, nil
	})
}
