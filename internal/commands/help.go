package commands

import (
	"consul-telegram-bot/internal/metrics"
	"consul-telegram-bot/internal/router"
)

func Help(c *router.Context) {
	metrics.TelegramCommandsProcessed.WithLabelValues("help", "success").Inc()

	baseHelp := "<b>Available Commands:</b>\n\n" +
		"<b>General:</b>\n" +
		"/id - Get chat ID.\n" +
		"/help - Get a list of commands.\n\n" +
		"<b>Ecosystem:</b>\n" +
		"/website - Get website link.\n" +
		"/ca - Get contract address.\n" +
		"/chart - View chart on Dexscreener.\n\n" +
		"<b>AI:</b>\n" +
		"/summary - Get AI summary of recent messages."

	if c.IsManager() {
		adminHelp := "\n\n<b>Admin:</b>\n" +
			"/setup - Setup wizard.\n" +
			"/set - Configure settings.\n" +
			"/clear - Clear all settings.\n" +
			"/define_thread_id - Set thread.\n" +
			"/retransmit - Broadcast message."
		c.SendAnswer(baseHelp + adminHelp)
	} else {
		c.SendAnswer(baseHelp)
	}
}
