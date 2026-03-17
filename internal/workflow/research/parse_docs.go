package research

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"ppt-smasher/internal/config"
	"ppt-smasher/internal/db"
)

// ParseDocs 模拟 MinerU 解析文档并使用多模态 Embedding 直接存入向量库
func ParseDocs(ctx context.Context, theme string, givenDocs []string) ([]string, []string, error) {
	log.Printf("[ResearchTeam:ParseDocs] 开始解析 %d 篇用户上传文档 (MinerU + Multimodal Embedding)...", len(givenDocs))

	var parsedTexts []string
	var imageIDs []string

	for i, docName := range givenDocs {
		// 修改：尝试从 docs/ 目录查找真实文件
		docPath := filepath.Join("docs", docName)
		if _, err := os.Stat(docPath); os.IsNotExist(err) {
			log.Printf("[ResearchTeam:ParseDocs] 找不到文档 %s，使用 Mock 数据", docPath)
			text := fmt.Sprintf("Docs[%d] Parsed Content: Based on %s ... (Simulated MinerU output)", i, docName)
			parsedTexts = append(parsedTexts, text)
			db.AddDocumentChunk(ctx, theme, fmt.Sprintf("%s-text-%d", docName, i), text)
			continue
		}

		log.Printf("[ResearchTeam:ParseDocs] 正处理真实文档: %s", docPath)

		// 集成 MinerU 真实调用
		markdown, imgURLs, err := ParseWithMinerU(ctx, docPath)
		if err != nil {
			log.Printf("[ResearchTeam:ParseDocs] MinerU 解析失败 %s: %v, 使用降级方案", docName, err)
			// 降级方案：原有的简单读取逻辑
			info, _ := os.Stat(docPath)
			text := fmt.Sprintf("Fallback Content of %s (Size: %d bytes). MinerU failed: %v", docName, info.Size(), err)
			parsedTexts = append(parsedTexts, text)
			db.AddDocumentChunk(ctx, theme, fmt.Sprintf("%s-text-%d", docName, i), text)
		} else {
			log.Printf("[ResearchTeam:ParseDocs] MinerU 解析成功: %s, 提取 %d 张图片", docName, len(imgURLs))
			parsedTexts = append(parsedTexts, markdown)
			db.AddDocumentChunk(ctx, theme, fmt.Sprintf("%s-mineru-%d", docName, i), markdown)

			// 下载 MinerU 提取的图片
			for j, imgURL := range imgURLs {
				imgID := fmt.Sprintf("%s-mineru-img-%d-%d", docName, i, j)
				tempDir := config.GlobalConfig.Paths.TempDir
				if tempDir == "" {
					tempDir = "temp"
				}
				imgDir := filepath.Join(tempDir, "images")
				savePath := filepath.Join(imgDir, imgID+".png")
				if err := DownloadImage(ctx, imgURL, savePath); err == nil {
					db.AddImageChunk(ctx, theme, imgID, imgURL, savePath)
					imageIDs = append(imageIDs, imgID)
				}
			}
		}

		// 检查 temp/images/ 下是否有同名图片作为解析结果的模拟
		imgID := fmt.Sprintf("%s-img-%d", docName, i)
		tempDir := config.GlobalConfig.Paths.TempDir
		if tempDir == "" {
			tempDir = "temp"
		}
		imgDir := filepath.Join(tempDir, "images")
		mockPath := filepath.Join(imgDir, imgID+".png")
		if _, err := os.Stat(mockPath); err == nil {
			log.Printf("[ResearchTeam:ParseDocs] 发现解析关联图片: %s", mockPath)
			db.AddImageChunk(ctx, theme, imgID, "data:image/png;base64,...", mockPath)
			imageIDs = append(imageIDs, imgID)
		}
	}

	return parsedTexts, imageIDs, nil
}
