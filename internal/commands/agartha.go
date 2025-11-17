package commands

import (
	"consul-telegram-bot/internal/metrics"
	"consul-telegram-bot/internal/router"
)

func Agartha(c *router.Context) {
	metrics.TelegramCommandsProcessed.WithLabelValues("agartha", "success").Inc()

	message := "<b>Agartha:</b>\n\n" +
		"Comprehensive trading signals bot. Tracks DEX market activity, token launches, and anomalies across multiple chains with advanced risk assessment.\n\n" +
		"Discover: " + c.Config.AgarthaBotURL

	c.SendAnswer(message)
}
