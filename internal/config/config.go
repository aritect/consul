package config

import (
	"os"
	"strconv"
)

type Config struct {
	ManagerId        int64
	TelegramBotToken string
	StorePath        string
	HeliusRpcURL     string

	ProjectName  string
	TokenTicker  string
	Description  string
	WebsiteURL   string
	TokenAddress string
	DexURL       string
	AxiomURL     string

	LLMProvider string
	LLMAPIKey   string
	LLMModel    string
}

func New() *Config {
	storePath := getEnvString("LEVELDB_STORE_PATH")
	if storePath == "" {
		storePath = "./data/store"
	}

	return &Config{
		TelegramBotToken: getEnvString("TELEGRAM_BOT_TOKEN"),
		ManagerId:        getEnvInt64("MANAGER_ID"),
		StorePath:        storePath,
		HeliusRpcURL:     getEnvString("HELIUS_RPC_URL"),

		ProjectName:  getEnvString("PROJECT_NAME"),
		TokenTicker:  getEnvString("TOKEN_TICKER"),
		Description:  getEnvString("DESCRIPTION"),
		WebsiteURL:   getEnvString("WEBSITE_URL"),
		TokenAddress: getEnvString("TOKEN_ADDRESS"),
		DexURL:       getEnvString("DEX_URL"),
		AxiomURL:     getEnvString("AXIOM_URL"),

		LLMProvider: getEnvString("LLM_PROVIDER"),
		LLMAPIKey:   getEnvString("LLM_API_KEY"),
		LLMModel:    getEnvString("LLM_MODEL"),
	}
}

func getEnvString(key string) string {
	return os.Getenv(key)
}

func getEnvInt64(key string) int64 {
	s := os.Getenv(key)
	number, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0
	}

	return int64(number)
}
