package content

import (
	"context"
	"fmt"
	"log"
	"ppt-smasher/internal/db"
	"strings"

	"github.com/cloudwego/eino/compose"
)

// NewRAGNode 创建一个 RAG 节点，用于从向量数据库检索相关文本并注入到状态中
func NewRAGNode() *compose.Lambda {
	return compose.InvokableLambda(func(ctx context.Context, s TeamContentState) (TeamContentState, error) {
		log.Printf("[ContentTeam:RAG] 正在为主题 '%s' 检索相关背景资料...", s.Theme)

		if !s.VDBStatus {
			log.Println("[ContentTeam:RAG] 向量数据库未就绪，跳过检索")
			return s, nil
		}

		// 构建查询词，结合主题和大纲（如果已存在）
		query := s.Theme
		if s.Outline != "" {
			query = fmt.Sprintf("%s %s", s.Theme, s.Outline)
		}

		// 从向量数据库检索
		retrieved, err := db.SearchDocument(ctx, s.Theme, query, 5)
		if err != nil {
			log.Printf("[ContentTeam:RAG] 检索文档失败: %v", err)
			return s, nil // 即使失败也继续，不阻塞流
		}

		if len(retrieved) > 0 {
			s.VDBContext = strings.Join(retrieved, "\n\n")
			log.Printf("[ContentTeam:RAG] 检索到 %d 条相关片段", len(retrieved))
		} else {
			log.Println("[ContentTeam:RAG] 未找到相关片段")
		}

		return s, nil
	})
}
