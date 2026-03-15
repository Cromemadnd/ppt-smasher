package render

import (
	"context"
	_ "embed"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"ppt-stasher-backend/internal/config"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

//go:embed prompts/coder.md
var coderPromptTemplate string

func parsePythonSnippet(text string) string {
	start := strings.Index(text, "```python")
	if start != -1 {
		text = text[start+9:]
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

func NewScriptCoderNode() *compose.Lambda {
	return compose.InvokableLambda(func(ctx context.Context, s TeamRenderState) (TeamRenderState, error) {
		log.Println("[RenderTeam:Coder] 生成 Python 代码渲染 PPT，并进行闭环执行...")

		if len(s.ContentDrafts) == 0 {
			log.Println("[RenderTeam:Coder] 没有内容草稿可供渲染")
			return s, nil
		}

		chatModel, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
			Model:   config.GlobalConfig.LLM.VisualModel,
			APIKey:  config.GlobalConfig.LLM.APIKey,
			BaseURL: config.GlobalConfig.LLM.BaseURL,
		})
		if err != nil {
			return s, fmt.Errorf("init model failed: %w", err)
		}

		finalDraft := s.ContentDrafts[len(s.ContentDrafts)-1]
		chatTpl := prompt.FromMessages(schema.FString, schema.UserMessage(coderPromptTemplate))
		messages, err := chatTpl.Format(ctx, map[string]any{
			"draft": finalDraft,
		})
		if err != nil {
			return s, fmt.Errorf("format prompt failed: %w", err)
		}

		resp, err := chatModel.Generate(ctx, messages)
		if err != nil {
			return s, fmt.Errorf("generate python code failed: %w", err)
		}

		pythonCode := parsePythonSnippet(resp.Content)
		if pythonCode == "" {
			pythonCode = resp.Content
		}

		// Save Python script locally to exec
		scriptPath := filepath.Join(os.TempDir(), "render_ppt.py")
		if err := os.WriteFile(scriptPath, []byte(pythonCode), 0644); err != nil {
			return s, fmt.Errorf("failed to save python script: %w", err)
		}

		log.Printf("[RenderTeam:Coder] 已生成渲染脚本保存在: %s。执行该脚本...", scriptPath)

		// 调用 OS Exec 执行 python 脚本
		cmd := exec.Command("python", scriptPath)
		output, err := cmd.CombinedOutput()
		if err != nil {
			log.Printf("[RenderTeam:Coder] 执行渲染脚本失败: %s\n%s", err.Error(), string(output))
			return s, fmt.Errorf("python exec failed: %v", err)
		}

		log.Println("[RenderTeam:Coder] PPTX 渲染成功生成 output.pptx")
		s.RenderResults = append(s.RenderResults, "output.pptx")

		return s, nil
	})
}
