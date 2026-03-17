package research

import (
	"context"
	"fmt"
	"log"

	"ppt-smasher/internal/config"
	// "github.com/cloudwego/eino-ext/components/model/openai"
	// "github.com/cloudwego/eino/components/prompt"
	// "github.com/cloudwego/eino/schema"
)

// ParseDocs 模拟 MinerU 解析文档并用 VLM (例如 Vision 型) 为文档图片生成描述
// 由于依赖外部 MinerU 和 VLM 视觉 API，这里我们提供一个简易的调用框架实现
func ParseDocs(ctx context.Context, givenDocs []string) ([]string, []string, error) {
	log.Printf("[ResearchTeam:ParseDocs] 开始解析 %d 篇用户上传文档 (MinerU + VLM)...", len(givenDocs))

	var parsedTexts []string
	var parsedImageDesc []string

	// 这里如果是真实的外部 MinerU 调用，应发起 HTTP/RPC 请求解析文档格式，提取图片并扔给 VLM
	// 此处模拟解析过程
	vlmModelID := config.GlobalConfig.LLM.ResearcherModel
	log.Printf("[ResearchTeam:ParseDocs] 使用 VLM模型: %s 生成图片描述", vlmModelID)

	for i, doc := range givenDocs {
		// Mock: 调用 MinerU 文本提取
		text := fmt.Sprintf("Docs[%d] Parsed Content: Based on %s ... (Simulated MinerU output)", i, doc)
		parsedTexts = append(parsedTexts, text)

		// Mock: 调用 VLM 获取图片描述
		imgDesc := fmt.Sprintf("Docs[%d] Image Description: A generated chart showing data... (Simulated VLM output)", i)
		parsedImageDesc = append(parsedImageDesc, imgDesc)
	}

	return parsedTexts, parsedImageDesc, nil
}
