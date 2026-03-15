package content

import (
	"context"
	"log"

	"github.com/cloudwego/eino/compose"
)

func NewOutlineDirectorNode() *compose.Lambda {
	return compose.InvokableLambda(func(ctx context.Context, s TeamContentState) (TeamContentState, error) {
		log.Println("[ContentTeam:Director] 起草幻灯片大纲，选择模板Schema...")
		return s, nil
	})
}
