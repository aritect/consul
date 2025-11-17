package commands

import (
	"consul-telegram-bot/internal/metrics"
	"consul-telegram-bot/internal/model"
	"consul-telegram-bot/internal/router"
)

func Start(c *router.Context) {
	chat := c.Message.Chat

	_, err := model.NewRecipient(chat.ID, model.RecipientType(chat.Type), c.Message.ThreadID)
	if err != nil {
		metrics.TelegramCommandsProcessed.WithLabelValues("start", "error").Inc()
		metrics.ErrorsTotal.WithLabelValues("command", "create_recipient").Inc()
		c.SendAnswer("🚧 Unfortunately, something went wrong.")
		return
	}

	metrics.TelegramCommandsProcessed.WithLabelValues("start", "success").Inc()
	c.SendAnswer(
		"<b>Consul</b> — your guide to the Aritect ecosystem.\n\n" +
			"I'm here to help you navigate resources, connect with intelligent bots, and access essential information about the Aritect platform.\n\n" +
			"<b>Available resources:</b>\n" +
			"- Platform information and links.\n" +
			"- Token contract address.\n" +
			"- Trading bots: Arbiter and Agartha.\n" +
			"- Real-time charts and analytics.\n\n" +
			"Use /help to explore all commands.",
	)
}
