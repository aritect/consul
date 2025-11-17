package commands

import (
	"consul-telegram-bot/internal/metrics"
	"consul-telegram-bot/internal/router"
)

func Chart(c *router.Context) {
	metrics.TelegramCommandsProcessed.WithLabelValues("chart", "success").Inc()

	message := "View on <a href=\"" + c.Config.ChartURL + "\">Dexscreener</a> for more details."

	c.SendAnswer(message)
}
