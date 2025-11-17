package commands

import (
	"consul-telegram-bot/internal/metrics"
	"consul-telegram-bot/internal/router"
	"fmt"
)

func Id(c *router.Context) {
	metrics.TelegramCommandsProcessed.WithLabelValues("id", "success").Inc()
	c.SendAnswer(fmt.Sprintf("Your chat ID: <b>%d</b>", c.Message.Chat.ID))
}
