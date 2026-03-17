package config

import (
	"log"
	"strings"

	"github.com/spf13/viper"
)

type VDBConfig struct {
	Type     string `yaml:"type"` // "postgres" or "milvus"
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DBName   string `yaml:"dbname"`  // For Postgres
	SSLMode  string `yaml:"sslmode"` // For Postgres
}

type SearchConfig struct {
	TavilyAPIKey string `mapstructure:"tavily_api_key"`
}

type MinerUConfig struct {
	APIKey  string `mapstructure:"api_key"`
	BaseURL string `mapstructure:"base_url"`
}

type Config struct {
	LLM    LLMConfig    `mapstructure:"llm"`
	Search SearchConfig `mapstructure:"search"`
	VDB    VDBConfig    `mapstructure:"vdb"`
	MinerU MinerUConfig `mapstructure:"mineru"`
	Paths  PathConfig   `mapstructure:"paths"`
}

type PathConfig struct {
	TempDir      string `mapstructure:"temp_dir"`
	MinerUResult string `mapstructure:"mineru_result"`
}

type LLMConfig struct {
	Provider          string `mapstructure:"provider"` // e.g. "openai", "ark"
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

func InitConfig(configPath []string) {
	// 加载 .env 文件（如果存在）
	// _ = godotenv.Load()

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	for _, path := range configPath {
		viper.AddConfigPath(path)
	}

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
