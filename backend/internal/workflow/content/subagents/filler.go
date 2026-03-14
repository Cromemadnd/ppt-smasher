package subagents

import (
	"context"
	"github.com/cloudwego/eino/compose"
	"log"
	"ppt-stasher-backend/internal/workflow/content"
)

func NewContentFillerNode() compose.InvokableLambda[content.TeamContentState, content.TeamContentState] {
	return compose.InvokableLambda(func(ctx context.Context, s content.TeamContentState) (content.TeamContentState, error) {
		log.Println("[ContentTeam:Director] 从 VDB 提取详细文案，填充入骨架...")
		s.FilledContentDraft = append(s.FilledContentDraft, "Draft data payload...")
		return s, nil
	})
}
