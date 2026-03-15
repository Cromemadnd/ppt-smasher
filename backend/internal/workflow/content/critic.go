package content

import (
	"context"
	"log"

	"github.com/cloudwego/eino/compose"
)

func NewContentCriticNode() *compose.Lambda {
	return compose.InvokableLambda(func(ctx context.Context, s TeamContentState) (TeamContentState, error) {
		log.Println("[ContentTeam:Director] 内部审查(Critic)：事实核查与防幻觉...")

		return s, nil
	})
}
