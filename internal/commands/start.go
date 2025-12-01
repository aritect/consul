package commands

import (
	"consul-telegram-bot/internal/metrics"
	"consul-telegram-bot/internal/model"
	"consul-telegram-bot/internal/router"
)

func Start(c *router.Context) {
	chat := c.Message.Chat

	recipient, err := model.NewRecipient(chat.ID, model.RecipientType(chat.Type), c.Message.ThreadID)
	if err != nil {
		metrics.TelegramCommandsProcessed.WithLabelValues("start", "error").Inc()
		metrics.ErrorsTotal.WithLabelValues("command", "create_recipient").Inc()
		c.SendAnswer("🚧 Unfortunately, something went wrong.")
		return
	}

	projectName := model.GetWithFallback(recipient.ProjectName, c.Config.ProjectName)
	if projectName == "" {
		projectName = "our community"
	}

	metrics.TelegramCommandsProcessed.WithLabelValues("start", "success").Inc()
	c.SendAnswer(
		"<b>Consul</b> — your AI-powered community assistant.\n\n" +
			"I'm here to help you navigate resources and access essential information about " + projectName + ".\n\n" +
			"<b>Available resources:</b>\n" +
			"- Platform information and links.\n" +
			"- Token contract address.\n" +
			"- Real-time charts and analytics.\n" +
			"- Buy notifications.\n\n" +
			"Use /help to explore all commands.",
	)
}
