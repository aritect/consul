package commands

import (
	"consul-telegram-bot/internal/metrics"
	"consul-telegram-bot/internal/router"
	"strings"
)

func Website(c *router.Context) {
	metrics.TelegramCommandsProcessed.WithLabelValues("website", "success").Inc()

	message :=
		"Aritect combines real-time blockchain analytics with adaptive AI to transform how you discover and evaluate investment opportunities across decentralized markets.\n\n" +
			"Link: <a href=\"" + c.Config.WebsiteURL + "\">" + strings.Split(c.Config.WebsiteURL, "//")[1] + "</a>"

	c.SendAnswer(message)
}
