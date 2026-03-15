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

//go:embed prompts/filler.md
var fillerPromptTemplate string

func parseJSONSnippetFiller(text string) string {
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

func NewContentFillerNode() *compose.Lambda {
	return compose.InvokableLambda(func(ctx context.Context, s TeamContentState) (TeamContentState, error) {
		log.Println("[ContentTeam:Filler] 根据大纲起草详细文案，并填充入具体的骨架...")

		chatModel, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
			Model:   config.GlobalConfig.LLM.ContentModel,
			APIKey:  config.GlobalConfig.LLM.APIKey,
			BaseURL: config.GlobalConfig.LLM.BaseURL,
		})
		if err != nil {
			return s, fmt.Errorf("init model failed: %w", err)
		}

		chatTpl := prompt.FromMessages(schema.FString, schema.UserMessage(fillerPromptTemplate))

		templatesStr := strings.Join(s.AvailableLayouts, "\n\n")

		krContext := "No Extracted Knowledge Provided here, rely on general domain knowledge."
		// Note: If VDBStatus is true, typically we'd fetch actual RAG knowledge here.
		// For now we pass a generic string or hook up to the RAG retrieval tool directly.

		messages, err := chatTpl.Format(ctx, map[string]any{
			"theme":   s.Theme,
			"outline": s.Outline,
			"context": krContext,
			"schemas": templatesStr,
		})
		if err != nil {
			return s, fmt.Errorf("format prompt failed: %w", err)
		}

		resp, err := chatModel.Generate(ctx, messages)
		if err != nil {
			return s, fmt.Errorf("generate content filler failed: %w", err)
		}

		fillerJSON := parseJSONSnippetFiller(resp.Content)
		if fillerJSON == "" { // fallback if no codeblock
			fillerJSON = resp.Content
		}

		// Save the finalized contents into Drafts.
		s.FilledContentDraft = append(s.FilledContentDraft, fillerJSON)
		log.Printf("[ContentTeam:Filler] 生成文案并填充完成: 产生 %d 个草案版本", len(s.FilledContentDraft))

		return s, nil
	})
}
