package content

import (
	"context"
	_ "embed"
	"fmt"
	"log"
	"strings"

	"ppt-stasher-backend/internal/config"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

//go:embed prompts/outline.md
var outlinePromptTemplate string

func parseJSONSnippetOutline(text string) string {
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

func NewOutlineDirectorNode() *compose.Lambda {
	return compose.InvokableLambda(func(ctx context.Context, s TeamContentState) (TeamContentState, error) {
		log.Println("[ContentTeam:Director] 起草幻灯片大纲，选择模板Schema...")

		chatModel, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
			Model:   config.GlobalConfig.LLM.ContentModel,
			APIKey:  config.GlobalConfig.LLM.APIKey,
			BaseURL: config.GlobalConfig.LLM.BaseURL,
		})
		if err != nil {
			return s, fmt.Errorf("init model failed: %w", err)
		}

		chatTpl := prompt.FromMessages(schema.FString, schema.UserMessage(outlinePromptTemplate))

		templatesStr := strings.Join(s.AvailableLayouts, "\n\n")

		kr := "Not Available"
		if s.VDBStatus {
			kr = "Available, knowledge context was gathered."
		}

		messages, err := chatTpl.Format(ctx, map[string]any{
			"theme":            s.Theme,
			"knowledge_ready":  kr,
			"schemas":          templatesStr,
			"current_feedback": s.CurrentFeedback,
		})
		if err != nil {
			return s, fmt.Errorf("format prompt failed: %w", err)
		}

		resp, err := chatModel.Generate(ctx, messages)
		if err != nil {
			return s, fmt.Errorf("generate outline failed: %w", err)
		}

		outlineJSON := parseJSONSnippetOutline(resp.Content)
		if outlineJSON == "" { // fallback if no codeblock
			outlineJSON = resp.Content
		}

		s.Outline = outlineJSON
		log.Printf("[ContentTeam:Director] 生成大纲完成:\n%s", s.Outline)

		return s, nil
	})
}
