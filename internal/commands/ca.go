package commands

import (
	"consul-telegram-bot/internal/metrics"
	"consul-telegram-bot/internal/model"
	"consul-telegram-bot/internal/router"
)

func CA(c *router.Context) {
	chat := c.Message.Chat
	recipient, _ := model.FindRecipient(chat.ID)

	var tokenAddress string
	if recipient != nil {
		tokenAddress = model.GetWithFallback(recipient.TokenAddress, c.Config.TokenAddress)
	} else {
		tokenAddress = c.Config.TokenAddress
	}

	if tokenAddress == "" {
		metrics.TelegramCommandsProcessed.WithLabelValues("ca", "error").Inc()
		c.SendAnswer("🚧 Token address is not configured. Use /setup or set TOKEN_ADDRESS env.")
		return
	}

	metrics.TelegramCommandsProcessed.WithLabelValues("ca", "success").Inc()

	message := "<code>" + tokenAddress + "</code>"

	c.SendAnswer(message)
}
