package config

import (
	"log"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	LLM LLMConfig `mapstructure:"llm"`
}

type LLMConfig struct {
	APIKey            string `mapstructure:"api_key"`
	BaseURL           string `mapstructure:"base_url"`
	BossModel         string `mapstructure:"boss_model"`
	ResearcherModel   string `mapstructure:"researcher_model"`
	ContentModel      string `mapstructure:"content_model"`
	VisualModel       string `mapstructure:"visual_model"`
	EmbeddingProvider string `mapstructure:"embedding_provider"` // "openai" is currently supported
	EmbeddingModel    string `mapstructure:"embedding_model"`
	EmbeddingDim      int    `mapstructure:"embedding_dim"`
}

var GlobalConfig *Config

func InitConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")        // 本地当前目录
	viper.AddConfigPath("./config") // 或 config 目录

	// 环境变量支持
	viper.AutomaticEnv()
	// 支持将形如 LLM_API_KEY 的环境变量映射到嵌套的 key
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Warning: Config file not found, trying environment variables: %v", err)
	}

	GlobalConfig = &Config{}
	if err := viper.Unmarshal(GlobalConfig); err != nil {
		log.Fatalf("Failed to unmarshal config: %v", err)
	}

	log.Println("Configuration loaded successfully.")
}
