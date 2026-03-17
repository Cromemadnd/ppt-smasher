package research

import (
	"context"
	"fmt"
	"log"

	"ppt-smasher/internal/db"
	// "github.com/cloudwego/eino-ext/components/model/openai"
)

// ParseDocs 模拟 MinerU 解析文档并使用多模态 Embedding 直接存入向量库
func ParseDocs(ctx context.Context, theme string, givenDocs []string) ([]string, []string, error) {
	log.Printf("[ResearchTeam:ParseDocs] 开始解析 %d 篇用户上传文档 (MinerU + Multimodal Embedding)...", len(givenDocs))

	var parsedTexts []string
	var imageIDs []string

	for i, doc := range givenDocs {
		// Mock: 调用 MinerU 文本提取
		text := fmt.Sprintf("Docs[%d] Parsed Content: Based on %s ... (Simulated MinerU output)", i, doc)
		parsedTexts = append(parsedTexts, text)

		// 存入文本向量
		db.AddDocumentChunk(ctx, theme, fmt.Sprintf("%s-text-%d", doc, i), text)

		// Mock: 提取图片并直接 Embedding
		// 在实际场景中，这里会得到图片的 base64 或路径
		mockBase64 := "data:image/jpeg;base64,/9j/4AAQSkZJRg..." 
		imgID := fmt.Sprintf("%s-img-%d", doc, i)
		// 模拟存入本地文件系统后的路径
		mockPath := fmt.Sprintf("images/%s.png", imgID)

		log.Printf("[ResearchTeam:ParseDocs] 直接对图片 %s 进行 Embedding 并记录路径 %s 到 Milvus", imgID, mockPath)
		err := db.AddImageChunk(ctx, theme, imgID, mockBase64, mockPath)
		if err != nil {
			log.Printf("[ResearchTeam:ParseDocs] 图片入库失败: %v", err)
		} else {
			imageIDs = append(imageIDs, imgID)
		}
	}

	return parsedTexts, imageIDs, nil
}
