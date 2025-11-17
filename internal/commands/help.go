package commands

import (
	"consul-telegram-bot/internal/metrics"
	"consul-telegram-bot/internal/router"
)

func Help(c *router.Context) {
	metrics.TelegramCommandsProcessed.WithLabelValues("help", "success").Inc()

	c.SendAnswer(
		"<b>Available Commands:</b>\n\n" +
			"<b>General:</b>\n" +
			"/id - Get chat ID.\n" +
			"/help - Get a list of commands.\n\n" +
			"<b>Ecosystem:</b>\n" +
			"/website - Get Aritect website link.\n" +
			"/ca - Get contract address.\n" +
			"/chart - View chart on Dexscreener.\n\n" +
			"<b>Bots:</b>\n" +
			"/arbiter - Discover arbitrage opportunities.\n" +
			"/agartha - Track market activity, token launches, and anomalies across multiple chains with advanced risk assessment.",
	)
}
