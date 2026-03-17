package llm

import (
	"context"
	"fmt"
	"log"

	"ppt-smasher/internal/config"

	"github.com/cloudwego/eino-ext/components/model/ark"
	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"
)

var (
	bossModel       model.ToolCallingChatModel
	researcherModel model.ToolCallingChatModel
	contentModel    model.ToolCallingChatModel
	visualModel     model.ToolCallingChatModel
)

func InitChatModels(ctx context.Context) {
	conf := config.GlobalConfig.LLM

	bossModel = mustGetChatModel(ctx, conf.BossModel)
	researcherModel = mustGetChatModel(ctx, conf.ResearcherModel)
	contentModel = mustGetChatModel(ctx, conf.ContentModel)
	visualModel = mustGetChatModel(ctx, conf.VisualModel)

	log.Println("LLM Chat Models initialized successfully.")
}

func GetBossModel() model.ToolCallingChatModel {
	return bossModel
}

func GetResearcherModel() model.ToolCallingChatModel {
	return researcherModel
}

func GetContentModel() model.ToolCallingChatModel {
	return contentModel
}

func GetVisualModel() model.ToolCallingChatModel {
	return visualModel
}

func mustGetChatModel(ctx context.Context, modelName string) model.ToolCallingChatModel {
	if modelName == "" {
		return nil
	}
	m, err := NewChatModel(ctx, modelName)
	if err != nil {
		log.Fatalf("failed to init chat model %s: %v", modelName, err)
	}
	return m
}

func NewChatModel(ctx context.Context, modelName string) (model.ToolCallingChatModel, error) {
	provider := config.GlobalConfig.LLM.Provider
	switch provider {
	case "ark":
		return ark.NewChatModel(ctx, &ark.ChatModelConfig{
			APIKey: config.GlobalConfig.LLM.APIKey,
			Model:  modelName,
			Region: "", // usually ark doesn't strictly need a custom region if endpoint is handled or not provided
		})
	case "openai", "":
		return openai.NewChatModel(ctx, &openai.ChatModelConfig{
			Model:   modelName,
			APIKey:  config.GlobalConfig.LLM.APIKey,
			BaseURL: config.GlobalConfig.LLM.BaseURL,
		})
	default:
		return nil, fmt.Errorf("unsupported LLM provider: %s", provider)
	}
}
