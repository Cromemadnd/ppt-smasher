package research

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

//go:embed prompts/research_leader.md
var leaderPromptTemplate string

type StudyDimensions struct {
	DocQueries   []string `json:"doc_queries"`
	ImageQueries []string `json:"image_queries"`
	DataQueries  []string `json:"data_queries"`
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

func NewResearchLeaderNode() *compose.Lambda {
	return compose.InvokableLambda(func(ctx context.Context, s TeamResearchState) (TeamResearchState, error) {
		log.Printf("[ResearchTeam:Leader] 为主题『%s』规划调研维度...", s.Theme)

		chatModel, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
			Model:   config.GlobalConfig.LLM.ResearcherModel,
			APIKey:  config.GlobalConfig.LLM.APIKey,
			BaseURL: config.GlobalConfig.LLM.BaseURL, // Assuming BaseURL is supported for OpenAI generic interface
		})
		if err != nil {
			log.Printf("[ResearchTeam:Leader] 初始化模型失败: %v。将使用默认 mock 数据。", err)
			return mockFallback(s), nil
		}

		chatTpl := prompt.FromMessages(schema.FString, schema.UserMessage(leaderPromptTemplate))
		messages, err := chatTpl.Format(ctx, map[string]any{
			"theme": s.Theme,
		})
		if err != nil {
			log.Printf("[ResearchTeam:Leader] 构建提示词失败: %v。将使用默认 mock 数据。", err)
			return mockFallback(s), nil
		}

		resp, err := chatModel.Generate(ctx, messages)
		if err != nil {
			log.Printf("[ResearchTeam:Leader] LLM 生成失败: %v。将使用默认 mock 数据。", err)
			return mockFallback(s), nil
		}

		content := parseJSONSnippet(resp.Content)
		var dims StudyDimensions
		if err := json.Unmarshal([]byte(content), &dims); err != nil {
			log.Printf("[ResearchTeam:Leader] 解析 LLM 响应失败: %v\nResp: %s\n将使用默认 mock 数据。", err, content)
			return mockFallback(s), nil
		}

		s.DocQueries = dims.DocQueries
		s.ImageQueries = dims.ImageQueries
		s.DataQueries = dims.DataQueries

		log.Printf("[ResearchTeam:Leader] 生成了 %d 个文档关键词, %d 个图片关键词, %d 个数据关键词",
			len(s.DocQueries), len(s.ImageQueries), len(s.DataQueries))

		return s, nil
	})
}

func mockFallback(s TeamResearchState) TeamResearchState {
	s.DocQueries = []string{
		fmt.Sprintf("%s 行业概况 深度分析", s.Theme),
		fmt.Sprintf("%s 基础知识 百科", s.Theme),
	}
	s.ImageQueries = []string{
		fmt.Sprintf("%s 高清大图 职场风格", s.Theme),
		fmt.Sprintf("%s 创意构图 抽象概念", s.Theme),
	}
	s.DataQueries = []string{
		fmt.Sprintf("%s 相关 统计数据 2024", s.Theme),
		fmt.Sprintf("%s 市场份额 增长趋势 图表", s.Theme),
	}
	return s
}

func BuildResearchTeamGraph() *compose.Graph[TeamResearchState, TeamResearchState] {
	g := compose.NewGraph[TeamResearchState, TeamResearchState]()
	modelID := config.GlobalConfig.LLM.ResearcherModel
	log.Printf("ResearchTeam Model initialized with %s", modelID)

	_ = g.AddLambdaNode("research_leader", NewResearchLeaderNode())
	_ = g.AddLambdaNode("search_document", NewSearchDocumentNode())
	_ = g.AddLambdaNode("search_image", NewSearchImageNode())
	_ = g.AddLambdaNode("search_analytics", NewSearchAnalyticsNode())

	_ = g.AddEdge(compose.START, "research_leader")
	_ = g.AddEdge("research_leader", "search_document")
	_ = g.AddEdge("search_document", "search_image")
	_ = g.AddEdge("search_image", "search_analytics")
	_ = g.AddEdge("search_analytics", compose.END)

	return g
}
