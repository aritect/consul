package config

import (
	"os"
	"strconv"
)

type Config struct {
	ManagerId        int64
	TelegramBotToken string
	StorePath        string
	WebsiteURL       string
	TokenAddress     string
	ArbiterBotURL    string
	AgarthaBotURL    string
	DexscreenerUrl   string
	AxiomUrl         string
	HeliusRpcURL     string
}

func New() *Config {
	return &Config{
		TelegramBotToken: getEnvString("TELEGRAM_BOT_TOKEN"),
		ManagerId:        getEnvInt64("MANAGER_ID"),
		StorePath:        getEnvString("LEVELDB_STORE_PATH"),
		WebsiteURL:       getEnvString("WEBSITE_URL"),
		TokenAddress:     getEnvString("TOKEN_ADDRESS"),
		ArbiterBotURL:    getEnvString("ARBITER_BOT_URL"),
		AgarthaBotURL:    getEnvString("AGARTHA_BOT_URL"),
		DexscreenerUrl:   getEnvString("DEXSCREENER_URL"),
		AxiomUrl:         getEnvString("AXIOM_URL"),
		HeliusRpcURL:     getEnvString("HELIUS_RPC_URL"),
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
