package commands

import (
	"consul-telegram-bot/internal/metrics"
	"consul-telegram-bot/internal/router"
)

func Arbiter(c *router.Context) {
	metrics.TelegramCommandsProcessed.WithLabelValues("arbiter", "success").Inc()

	message := "<b>Arbiter:</b>\n\n" +
		"Specialized bot for arbitrage opportunities. Monitors CEX spread signals across multiple exchanges and chains, delivering instant alerts for profitable trades.\n\n" +
		"Discover: " + c.Config.ArbiterBotURL

	c.SendAnswer(message)
}
