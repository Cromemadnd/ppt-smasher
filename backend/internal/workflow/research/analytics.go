package research

import (
	"context"
	"log"

	"github.com/cloudwego/eino/compose"
)

func NewSearchAnalyticsNode() *compose.Lambda {
	return compose.InvokableLambda(func(ctx context.Context, s TeamResearchState) (TeamResearchState, error) {
		log.Println("[ResearchTeam] 正在检索数据图表...")
		s.VDBStatus = true
		return s, nil
	})
}
