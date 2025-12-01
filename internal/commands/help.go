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
			"/website - Get website link.\n" +
			"/ca - Get contract address.\n" +
			"/chart - View chart on Dexscreener.\n\n" +
			"<b>Configuration:</b>\n" +
			"/setup - Setup wizard.\n" +
			"/set - Configure settings (admin).",
	)
}
