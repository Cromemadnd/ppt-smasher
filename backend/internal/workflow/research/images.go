package research

import (
	"context"
	"log"

	"github.com/cloudwego/eino/compose"
)

func NewSearchImageNode() *compose.Lambda {
	return compose.InvokableLambda(func(ctx context.Context, s TeamResearchState) (TeamResearchState, error) {
		log.Println("[ResearchTeam] 正在检索配图与视觉素材...")
		return s, nil
	})
}
