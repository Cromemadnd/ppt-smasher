package content

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"ppt-stasher-backend/internal/config"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

//go:embed prompts/critic.md
var criticPromptTemplate string

type CriticResult struct {
	Decision       string `json:"decision"`
	Feedback       string `json:"feedback"`
	CorrectedDraft any    `json:"corrected_draft"`
}

func parseJSONSnippetCritic(text string) string {
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

func NewContentCriticNode() *compose.Lambda {
	return compose.InvokableLambda(func(ctx context.Context, s TeamContentState) (TeamContentState, error) {
		log.Println("[ContentTeam:Critic] 内部审查: 审核生成的幻灯片文案...")

		chatModel, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
			Model:   config.GlobalConfig.LLM.ContentModel,
			APIKey:  config.GlobalConfig.LLM.APIKey,
			BaseURL: config.GlobalConfig.LLM.BaseURL,
		})
		if err != nil {
			return s, fmt.Errorf("init model failed: %w", err)
		}

		if len(s.FilledContentDraft) == 0 {
			log.Println("[ContentTeam:Critic] 错误: 没有可以审查的草稿！")
			return s, nil
		}

		draftToReview := s.FilledContentDraft[len(s.FilledContentDraft)-1]

		chatTpl := prompt.FromMessages(schema.FString, schema.UserMessage(criticPromptTemplate))
		messages, err := chatTpl.Format(ctx, map[string]any{
			"theme":   s.Theme,
			"context": s.VDBContext,
			"outline": s.Outline,
			"draft":   draftToReview,
		})
		if err != nil {
			return s, fmt.Errorf("format prompt failed: %w", err)
		}

		resp, err := chatModel.Generate(ctx, messages)
		if err != nil {
			return s, fmt.Errorf("generate content critic failed: %w", err)
		}

		finalJSON := parseJSONSnippetCritic(resp.Content)
		if finalJSON == "" { // fallback if no codeblock
			finalJSON = resp.Content
		}

		var res CriticResult
		err = json.Unmarshal([]byte(finalJSON), &res)
		if err != nil {
			log.Printf("[ContentTeam:Critic] Default to Pass due to json parse err: %v", err)
			s.CriticDecision = "Pass"
		} else {
			s.CriticDecision = res.Decision
			s.CriticFeedback = res.Feedback
			s.CurrentFeedback = res.Feedback // Pass it along for next iterations
			b, _ := json.Marshal(res.CorrectedDraft)
			// Replace the last draft with the final revised version.
			s.FilledContentDraft[len(s.FilledContentDraft)-1] = string(b)
		}

		log.Printf("[ContentTeam:Critic] 内容审查结束: 决议=%s", s.CriticDecision)

		return s, nil
	})
}
