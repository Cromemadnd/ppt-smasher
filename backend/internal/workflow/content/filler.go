package content

import (
	"context"
	"log"

	"github.com/cloudwego/eino/compose"
)

func NewContentFillerNode() *compose.Lambda {
	return compose.InvokableLambda(func(ctx context.Context, s TeamContentState) (TeamContentState, error) {
		log.Println("[ContentTeam:Director] 从 VDB 提取详细文案，填充入骨架...")
		return s, nil
	})
}
