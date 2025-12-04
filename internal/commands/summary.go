package commands

import (
	"consul-telegram-bot/internal/llm"
	"consul-telegram-bot/internal/metrics"
	"consul-telegram-bot/internal/model"
	"consul-telegram-bot/internal/router"
	"consul-telegram-bot/internal/summarizer"
	"fmt"
	"sync"
	"time"
)

const (
	minMessagesForSummary     = 10
	defaultMessagesForSummary = 100
	summaryCooldown           = 60 * time.Second
)

var (
	summaryMu       sync.Mutex
	lastSummaryTime time.Time
)

func Summary(c *router.Context) {
	summaryMu.Lock()
	if time.Since(lastSummaryTime) < summaryCooldown {
		remaining := summaryCooldown - time.Since(lastSummaryTime)
		summaryMu.Unlock()
		c.SendAnswer(fmt.Sprintf("⏳ Please wait %d seconds before requesting another summary.", int(remaining.Seconds())))
		return
	}
	lastSummaryTime = time.Now()
	summaryMu.Unlock()

	if c.Config.LLMAPIKey == "" {
		metrics.TelegramCommandsProcessed.WithLabelValues("summary", "error").Inc()
		c.SendAnswer("🚧 Summary feature is not configured. Please set LLM_API_KEY.")
		return
	}

	chatID := c.Message.Chat.ID

	messageCount := model.CountMessages(chatID)
	if messageCount < minMessagesForSummary {
		metrics.TelegramCommandsProcessed.WithLabelValues("summary", "error").Inc()
		c.SendAnswer("⏳ Not enough messages for summary. Please wait for more community activity.")
		return
	}

	c.SendAnswer("✨ Generating summary, please wait...")

	messages, err := model.GetMessagesForSummary(chatID, defaultMessagesForSummary)
	if err != nil {
		c.Logger.Error("failed to get messages for summary: %s", err)
		metrics.TelegramCommandsProcessed.WithLabelValues("summary", "error").Inc()
		c.SendAnswer("🚧 Failed to retrieve messages.")
		return
	}

	provider, ok := llm.ParseProvider(c.Config.LLMProvider)
	if !ok {
		provider = llm.ProviderGroq
	}

	client := llm.NewClient(provider, c.Config.LLMAPIKey, c.Config.LLMModel)
	sum := summarizer.New(client)

	projectName := ""
	recipient, err := model.FindRecipient(chatID)
	if err == nil && recipient.ProjectName != "" {
		projectName = recipient.ProjectName
	} else if c.Config.ProjectName != "" {
		projectName = c.Config.ProjectName
	}

	summary, err := sum.GenerateSummary(messages, projectName)
	if err != nil {
		c.Logger.Error("failed to generate summary: %s", err)
		metrics.TelegramCommandsProcessed.WithLabelValues("summary", "error").Inc()
		c.SendAnswer("🚧 Failed to generate summary. Please try again later.")
		return
	}

	header := fmt.Sprintf("⚡️ Community summary based on last %d messages:\n\n", len(messages))
	c.SendAnswer(header + summary)

	metrics.TelegramCommandsProcessed.WithLabelValues("summary", "success").Inc()
}
