package commands

import (
	"consul-telegram-bot/internal/metrics"
	"consul-telegram-bot/internal/model"
	"consul-telegram-bot/internal/router"
)

func Chart(c *router.Context) {
	chat := c.Message.Chat
	recipient, _ := model.FindRecipient(chat.ID)

	var dexURL string
	if recipient != nil {
		dexURL = model.GetWithFallback(recipient.DexURL, c.Config.DexURL)
	} else {
		dexURL = c.Config.DexURL
	}

	if dexURL == "" {
		metrics.TelegramCommandsProcessed.WithLabelValues("chart", "error").Inc()
		c.SendAnswer("🚧 Dexscreener URL is not configured. Use /setup or set DEX_URL env.")
		return
	}

	metrics.TelegramCommandsProcessed.WithLabelValues("chart", "success").Inc()

	message := "View on <a href=\"" + dexURL + "\">Dexscreener</a> for more details."

	c.SendAnswer(message)
}
