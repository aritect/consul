package commands

import (
	"consul-telegram-bot/internal/metrics"
	"consul-telegram-bot/internal/router"
)

func NotFound(c *router.Context) {
	metrics.TelegramCommandsProcessed.WithLabelValues("not_found", "success").Inc()
	c.SendAnswer("💁 Sorry, I didn't understand your message.")
}
