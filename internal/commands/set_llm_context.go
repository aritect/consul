package commands

import (
	"consul-telegram-bot/internal/metrics"
	"consul-telegram-bot/internal/middlewares"
	"consul-telegram-bot/internal/model"
	"consul-telegram-bot/internal/router"
	"io"
	"strings"
)

func SetLLMContext(c *router.Context) {
	middlewares.Manager(setLLMContextHandler, c.Config.ManagerId)(c)
}

func setLLMContextHandler(c *router.Context) {
	var context string

	if c.Message.Document != nil {
		if !strings.HasSuffix(strings.ToLower(c.Message.Document.FileName), ".txt") {
			metrics.TelegramCommandsProcessed.WithLabelValues("set_llm_context", "error").Inc()
			c.SendAnswer("🚧 Please provide a .txt file.")
			return
		}

		fileContent, err := downloadFile(c, c.Message.Document.FileID)
		if err != nil {
			c.Logger.Error("failed to download file: %s", err)
			metrics.TelegramCommandsProcessed.WithLabelValues("set_llm_context", "error").Inc()
			c.SendAnswer("🚧 Failed to download file.")
			return
		}
		context = fileContent
	} else {
		if len(c.Args) == 0 {
			metrics.TelegramCommandsProcessed.WithLabelValues("set_llm_context", "error").Inc()
			c.SendAnswer("🚧 Please provide a context for the LLM or attach a .txt file.\n\nExample:\n/set_llm_context You are an expert in Aritect platform. Aritect is building trust infrastructure for ...")
			return
		}

		context = c.GetArgStringWithNewlines()
		if context == "" {
			metrics.TelegramCommandsProcessed.WithLabelValues("set_llm_context", "error").Inc()
			c.SendAnswer("🚧 Context cannot be empty.")
			return
		}
	}

	chatID := c.Message.Chat.ID
	_, err := model.NewLLMContext(chatID, context)
	if err != nil {
		c.Logger.Error("failed to save LLM context: %s", err)
		metrics.TelegramCommandsProcessed.WithLabelValues("set_llm_context", "error").Inc()
		c.SendAnswer("🚧 Failed to save LLM context.")
		return
	}

	metrics.TelegramCommandsProcessed.WithLabelValues("set_llm_context", "success").Inc()
	c.SendAnswer("✅ LLM context has been set successfully.")
}

func downloadFile(c *router.Context, fileID string) (string, error) {
	file, err := c.Bot.Bot.FileByID(fileID)
	if err != nil {
		return "", err
	}

	reader, err := c.Bot.Bot.File(&file)
	if err != nil {
		return "", err
	}
	defer reader.Close()

	content, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}

	return string(content), nil
}
