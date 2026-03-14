package subagents

import (
	"context"
	"github.com/cloudwego/eino/compose"
	"log"
	"ppt-stasher-backend/internal/workflow/research"
)

func NewSearchImageNode() compose.InvokableLambda[research.TeamResearchState, research.TeamResearchState] {
	return compose.InvokableLambda(func(ctx context.Context, s research.TeamResearchState) (research.TeamResearchState, error) {
		log.Println("[ResearchTeam] 正在检索配图与视觉素材...")
		return s, nil
	})
}
