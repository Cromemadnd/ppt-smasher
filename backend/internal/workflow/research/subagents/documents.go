package subagents

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/cloudwego/eino/compose"
	"ppt-stasher-backend/internal/db"
	"ppt-stasher-backend/internal/workflow/research"
)

func NewSearchDocumentNode() compose.InvokableLambda[research.TeamResearchState, research.TeamResearchState] {
	return compose.InvokableLambda(func(ctx context.Context, s research.TeamResearchState) (research.TeamResearchState, error) {
		log.Println("[ResearchTeam] 正在检索文本资料...")

		// 1. 将提供的参考资料存入 LanceDB 向量库 (模拟索引过程)
		for i, doc := range s.GivenDocuments {
			docID := fmt.Sprintf("doc_%d_%d", i, time.Now().UnixNano())
			log.Printf("[ResearchTeam:LanceDB] Indexing doc chunk: %s", docID)
			_ = db.AddDocumentChunk(ctx, docID, doc)
		}

		// 2. 将主题相关的要求在库中检索 (RAG)
		results, err := db.SearchDocument(ctx, s.Theme, 3)
		if err == nil && len(results) > 0 {
			log.Printf("[ResearchTeam:LanceDB] Retrieved %d chunks relating to theme", len(results))
			// 在真实的 RAG 中，这里会将结果加入 State 供后续 Agent 读取
			log.Printf("Context 1: %s", results[0])
		}

		s.VDBStatus = true
		return s, nil
	})
}
