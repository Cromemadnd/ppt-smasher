package research

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/cloudwego/eino/compose"
)

func NewSearchDocumentNode() *compose.Lambda {
	return compose.InvokableLambda(func(ctx context.Context, s TeamResearchState) (TeamResearchState, error) {
		log.Printf("[ResearchTeam:Worker] 正在并行搜索文档 (总量: %d)...", len(s.DocQueries))

		ch := make(chan string, len(s.DocQueries)*5)
		var wg sync.WaitGroup

		for _, q := range s.DocQueries {
			wg.Add(1)
			go func(query string) {
				defer wg.Done()
				res, err := SearchTavily(ctx, query, false)
				if err != nil {
					log.Printf("SearchTavily Document error for '%s': %v", query, err)
					return
				}
				for _, r := range res.Results {
					ch <- fmt.Sprintf("Source: %s\nTitle: %s\nContent: %s\n", r.URL, r.Title, r.Content)
				}
			}(q)
		}

		wg.Wait()
		close(ch)

		for content := range ch {
			s.Documents = append(s.Documents, content)
		}

		// Optionally parse `s.GivenDocuments` with MinerU later if needed
		s.VDBStatus = true // 标记知识库就绪
		return s, nil
	})
}

func NewSearchImageNode() *compose.Lambda {
	return compose.InvokableLambda(func(ctx context.Context, s TeamResearchState) (TeamResearchState, error) {
		log.Printf("[ResearchTeam:Worker] 正在并行搜索图片 (总量: %d)...", len(s.ImageQueries))

		ch := make(chan string, len(s.ImageQueries)*5)
		var wg sync.WaitGroup

		for _, q := range s.ImageQueries {
			wg.Add(1)
			go func(query string) {
				defer wg.Done()
				res, err := SearchTavily(ctx, query, true) // include images
				if err != nil {
					log.Printf("SearchTavily Image error for '%s': %v", query, err)
					return
				}
				for _, img := range res.Images {
					ch <- img.URL
				}
			}(q)
		}

		wg.Wait()
		close(ch)

		for url := range ch {
			s.Images = append(s.Images, url)
		}

		return s, nil
	})
}

func NewSearchAnalyticsNode() *compose.Lambda {
	return compose.InvokableLambda(func(ctx context.Context, s TeamResearchState) (TeamResearchState, error) {
		log.Printf("[ResearchTeam:Worker] 正在并行提取统计数据 (总量: %d)...", len(s.DataQueries))

		ch := make(chan string, len(s.DataQueries)*5)
		var wg sync.WaitGroup

		for _, q := range s.DataQueries {
			wg.Add(1)
			go func(query string) {
				defer wg.Done()
				res, err := SearchTavily(ctx, query, false)
				if err != nil {
					log.Printf("SearchTavily Analytics error for '%s': %v", query, err)
					return
				}
				for _, r := range res.Results {
					ch <- fmt.Sprintf("Source: %s\nTitle: %s\nContent: %s\n", r.URL, r.Title, r.Content)
				}
			}(q)
		}

		wg.Wait()
		close(ch)

		for content := range ch {
			s.Analytics = append(s.Analytics, content)
		}

		return s, nil
	})
}
