package research

import (
	"context"
	"fmt"
	"log"
	"ppt-smasher/internal/db"
	"sync"

	"github.com/cloudwego/eino/compose"
	"github.com/google/uuid"
)

func NewParallelTasksNode() *compose.Lambda {
	return compose.InvokableLambda(func(ctx context.Context, s TeamResearchState) (TeamResearchState, error) {
		log.Printf("[ResearchTeam:ParallelTasks] 开启并行执行网络搜集和文档解析...")

		var wg sync.WaitGroup
		var mu sync.Mutex

		// 1. 文档联网搜索
		wg.Add(1)
		go func() {
			defer wg.Done()
			docs := searchDocuments(ctx, s.DocQueries)
			mu.Lock()
			s.Documents = append(s.Documents, docs...)
			mu.Unlock()
		}()

		// 2. 图片联网搜索/向量库检索
		wg.Add(1)
		go func() {
			defer wg.Done()
			imgs := searchImages(ctx, s.ImageQueries)
			
			// 新增：从向量库检索图片描述
			for _, q := range s.ImageQueries {
				vdbImgs, err := db.SearchImage(ctx, s.Theme, q, 3)
				if err == nil {
					imgs = append(imgs, vdbImgs...)
				}
			}

			mu.Lock()
			s.Images = append(s.Images, imgs...)
			mu.Unlock()
		}()

		// 3. 数据维度的联网搜索
		wg.Add(1)
		go func() {
			defer wg.Done()
			analytics := searchAnalytics(ctx, s.DataQueries)
			mu.Lock()
			s.Analytics = append(s.Analytics, analytics...)
			mu.Unlock()
		}()

		// 4. 解析用户上传的参考资料 (MinerU + 多模态 Embedding)
		if len(s.GivenDocuments) > 0 {
			wg.Add(1)
			go func() {
				defer wg.Done()
				parseDocs, _, err := ParseDocs(ctx, s.Theme, s.GivenDocuments)
				if err != nil {
					log.Printf("[ResearchTeam:ParseDocs] 解析出错: %v", err)
					return
				}
				mu.Lock()
				s.Documents = append(s.Documents, parseDocs...)
				mu.Unlock()
			}()
		}

		wg.Wait()
		log.Printf("[ResearchTeam:ParallelTasks] 并行搜集完成。共汇总 %d 文档, %d 图片, %d 数据源.", len(s.Documents), len(s.Images), len(s.Analytics))
		return s, nil
	})
}

func searchDocuments(ctx context.Context, queries []string) []string {
	if len(queries) == 0 {
		return nil
	}
	ch := make(chan string, len(queries)*5)
	var wg sync.WaitGroup
	for _, q := range queries {
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
	var docs []string
	for content := range ch {
		docs = append(docs, content)
	}
	return docs
}

func searchImages(ctx context.Context, queries []string) []string {
	if len(queries) == 0 {
		return nil
	}
	ch := make(chan string, len(queries)*5)
	var wg sync.WaitGroup
	for _, q := range queries {
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
	var images []string
	for url := range ch {
		images = append(images, url)
	}
	return images
}

func searchAnalytics(ctx context.Context, queries []string) []string {
	if len(queries) == 0 {
		return nil
	}
	ch := make(chan string, len(queries)*5)
	var wg sync.WaitGroup
	for _, q := range queries {
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
	var analytics []string
	for content := range ch {
		analytics = append(analytics, content)
	}
	return analytics
}

func NewIndexVDBNode() *compose.Lambda {
	return compose.InvokableLambda(func(ctx context.Context, s TeamResearchState) (TeamResearchState, error) {
		log.Printf("[ResearchTeam:Index] 开始数据清洗并存入 VDB (LanceDB)...")

		// 收集所有文本数据，这里演示将 Documents 和 Analytics 全量分块
		allTexts := append(s.Documents, s.Analytics...)

		// 用来简单 Chunk (比如按回车/长度分块)，此例中一个 Text 就是一个源片段
		chunkCount := 0
		var wg sync.WaitGroup

		for _, content := range allTexts {
			if content == "" {
				continue
			}

			wg.Add(1)
			go func(text string) {
				defer wg.Done()
				// 清洗操作：如去除不需要的空白，规范化格式等，此略...
				chunkID := uuid.New().String()

				// 并行执行向量化写库
				err := db.AddDocumentChunk(ctx, s.Theme, chunkID, text)
				if err != nil {
					log.Printf("[ResearchTeam:Index] 写入 db 失败 (chunk %s): %v", chunkID[:6], err)
				}
			}(content)
			chunkCount++
		}

		wg.Wait()
		log.Printf("[ResearchTeam:Index] 共处理并尝试写入 %d 个 Chunk。", chunkCount)

		s.VDBStatus = true
		return s, nil
	})
}
