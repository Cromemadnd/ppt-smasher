package template

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"ppt-smasher/internal/config"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

//go:embed prompts/cluster.md
var clusterPromptTemplate string

type LayoutCategory struct {
	LayoutName string `json:"layout_name"`
	Category   string `json:"category"` // "structural" or "content"
}

type ClusterResponse struct {
	Layouts []LayoutCategory `json:"layouts"`
}

func parseClusterJSONSnippet(text string) string {
	start := strings.Index(text, "```json")
	if start != -1 {
		text = text[start+7:]
		end := strings.Index(text, "```")
		if end != -1 {
			text = text[:end]
		}
	} else {
		start = strings.Index(text, "```")
		if start != -1 {
			text = text[start+3:]
			end := strings.Index(text, "```")
			if end != -1 {
				text = text[:end]
			}
		}
	}
	return strings.TrimSpace(text)
}

func NewClusterLayoutNode() *compose.Lambda {
	return compose.InvokableLambda(func(ctx context.Context, s TeamTemplateState) (TeamTemplateState, error) {
		log.Println("[TemplateAnalyst:Cluster] 分析幻灯片功能分类，区分“结构页”与“内容页”...")

		if len(s.ExtractedStyle) == 0 {
			return s, nil
		}

		chatModel, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
			Model:   config.GlobalConfig.LLM.ContentModel,
			APIKey:  config.GlobalConfig.LLM.APIKey,
			BaseURL: config.GlobalConfig.LLM.BaseURL,
		})
		if err != nil {
			return s, fmt.Errorf("failed to init chat model: %v", err)
		}

		// Prepare payload containing all current extracted styles
		payloadBytes, _ := json.MarshalIndent(s.ExtractedStyle, "", "  ")

		chatTpl := prompt.FromMessages(schema.FString, schema.UserMessage(clusterPromptTemplate))
		messages, err := chatTpl.Format(ctx, map[string]any{
			"layouts": string(payloadBytes),
		})
		if err != nil {
			return s, fmt.Errorf("failed to format prompt: %v", err)
		}

		resp, err := chatModel.Generate(ctx, messages)
		if err != nil {
			return s, fmt.Errorf("failed to call LLM: %v", err)
		}

		content := parseClusterJSONSnippet(resp.Content)
		if content == "" {
			// If parsing fails completely, just use content.
			content = resp.Content
		}

		var clusterResp ClusterResponse
		if err := json.Unmarshal([]byte(content), &clusterResp); err != nil {
			log.Printf("[TemplateAnalyst:Cluster] JSON 解析失败, 返回透传状态: %v\nLLM Output: %s", err, content)
			return s, nil
		}

		categoryMap := make(map[string]string)
		for _, layout := range clusterResp.Layouts {
			categoryMap[layout.LayoutName] = layout.Category
		}

		for i, style := range s.ExtractedStyle {
			if cat, ok := categoryMap[style.LayoutName]; ok {
				s.ExtractedStyle[i].Category = cat
			} else {
				s.ExtractedStyle[i].Category = "content" // default
			}
		}

		log.Println("[TemplateAnalyst:Cluster] 分类完成！")
		return s, nil
	})
}
