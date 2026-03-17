package template

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"ppt-smasher/internal/llm"

	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

//go:embed prompts/schema.md
var schemaExtractorPromptTemplate string

type Element struct {
	Name  string `json:"name"`
	Type  string `json:"type"`
	ID    int    `json:"id,omitempty"`
	Index uint32 `json:"index,omitempty"`
}

type ExtractedSchema struct {
	Elements []Element `json:"elements"`
}

func parseJSONSnippet(text string) string {
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

func ExtractSchemaWithLLM(ctx context.Context, htmlView SlideHTMLSchema) (SlideLayoutSchema, error) {
	chatModel := llm.GetVisualModel()
	if chatModel == nil {
		return SlideLayoutSchema{LayoutName: htmlView.LayoutName}, fmt.Errorf("visual model not initialized")
	}

	chatTpl := prompt.FromMessages(schema.FString, schema.UserMessage(schemaExtractorPromptTemplate))
	messages, err := chatTpl.Format(ctx, map[string]any{
		"html": htmlView.HTML,
	})
	if err != nil {
		return SlideLayoutSchema{LayoutName: htmlView.LayoutName}, err
	}

	resp, err := chatModel.Generate(ctx, messages)
	if err != nil {
		return SlideLayoutSchema{LayoutName: htmlView.LayoutName}, err
	}

	content := parseJSONSnippet(resp.Content)
	var es ExtractedSchema
	if err := json.Unmarshal([]byte(content), &es); err != nil {
		return SlideLayoutSchema{LayoutName: htmlView.LayoutName}, fmt.Errorf("failed to decode LLM response: %v", err)
	}

	layoutSchema := SlideLayoutSchema{
		LayoutName:   htmlView.LayoutName,
		Placeholders: make([]Placeholder, 0),
	}

	for _, el := range es.Elements {
		layoutSchema.Placeholders = append(layoutSchema.Placeholders, Placeholder{
			ID:    el.ID,
			Type:  el.Type, // Can also be overwritten by Element.Name dynamically if we expose Semantic Type to downstream
			Index: el.Index,
			Name:  el.Name,
		})
	}
	return layoutSchema, nil
}

func NewSchemaExtractorNode() *compose.Lambda {
	return compose.InvokableLambda(func(ctx context.Context, s TeamTemplateState) (TeamTemplateState, error) {
		log.Println("[TemplateAnalyst] 正在利用 LLM 提取 HTML 的结构语义...")

		if len(s.HTMLViews) > 0 {
			var schemas []SlideLayoutSchema
			for _, view := range s.HTMLViews {
				ls, err := ExtractSchemaWithLLM(ctx, view)
				if err != nil {
					log.Printf("[TemplateAnalyst] 提取版式 '%s' 失败: %v。跳过...", view.LayoutName, err)
					continue
				}
				schemas = append(schemas, ls)
			}

			s.ExtractedStyle = schemas
			for _, sch := range schemas {
				b, _ := json.Marshal(sch)
				s.Schemas = append(s.Schemas, string(b))
			}
			log.Printf("[TemplateAnalyst] 基于 HTML 成功提取了 %d 种幻灯片语义化布局", len(schemas))
		}

		return s, nil
	})
}
