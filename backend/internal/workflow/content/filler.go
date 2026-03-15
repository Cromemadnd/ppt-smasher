package content

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"ppt-stasher-backend/internal/config"
	"ppt-stasher-backend/internal/db"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

//go:embed prompts/filler.md
var fillerPromptTemplate string

type FillerResult struct {
	Status               string   `json:"status"`
	NeedsResearchQueries []string `json:"needs_research_queries"`
	Slides               any      `json:"slides"`
}

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
		if s.VDBStatus {
			// Query LanceDB
			retrieved, err := db.SearchDocument(ctx, s.Theme, s.Theme+" "+s.Outline, 5)
			if err != nil {
				log.Printf("[ContentTeam:Filler] Failed to retrieve documents: %v", err)
			} else {
				krContext = strings.Join(retrieved, "\n\n")
				s.VDBContext = krContext // Store for Critic to use
			}
		}

		messages, err := chatTpl.Format(ctx, map[string]any{
			"theme":            s.Theme,
			"outline":          s.Outline,
			"context":          krContext,
			"schemas":          templatesStr,
			"current_feedback": s.CurrentFeedback,
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

		var res FillerResult
		err = json.Unmarshal([]byte(fillerJSON), &res)
		if err != nil {
			log.Printf("[ContentTeam:Filler] Default to Success due to json parse err: %v", err)
			s.FillerResultState = "Success"
			s.FilledContentDraft = append(s.FilledContentDraft, fillerJSON)
		} else {
			s.FillerResultState = res.Status
			s.ResearchQueries = res.NeedsResearchQueries
			b, _ := json.Marshal(res.Slides)
			s.FilledContentDraft = append(s.FilledContentDraft, string(b))
		}

		log.Printf("[ContentTeam:Filler] 生成文案并填充完成: 状态=%s, 草案数=%d", s.FillerResultState, len(s.FilledContentDraft))

		return s, nil
	})
}
