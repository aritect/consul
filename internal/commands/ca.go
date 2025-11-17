package commands

import (
	"consul-telegram-bot/internal/metrics"
	"consul-telegram-bot/internal/router"
)

func CA(c *router.Context) {
	metrics.TelegramCommandsProcessed.WithLabelValues("ca", "success").Inc()

	message := "<code>" + c.Config.TokenAddress + "</code>"

	c.SendAnswer(message)
}
