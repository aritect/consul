package commands

import (
	"consul-telegram-bot/internal/metrics"
	"consul-telegram-bot/internal/model"
	"consul-telegram-bot/internal/router"
	"strings"
)

func Website(c *router.Context) {
	chat := c.Message.Chat
	recipient, _ := model.FindRecipient(chat.ID)

	var websiteURL, description string
	if recipient != nil {
		websiteURL = model.GetWithFallback(recipient.WebsiteURL, c.Config.WebsiteURL)
		description = model.GetWithFallback(recipient.Description, c.Config.Description)
	} else {
		websiteURL = c.Config.WebsiteURL
		description = c.Config.Description
	}

	if websiteURL == "" {
		metrics.TelegramCommandsProcessed.WithLabelValues("website", "error").Inc()
		c.SendAnswer("🚧 Website URL is not configured. Use /setup or set WEBSITE_URL env.")
		return
	}

	metrics.TelegramCommandsProcessed.WithLabelValues("website", "success").Inc()

	descText := ""
	if description != "" {
		descText = description + "\n\n"
	}

	urlDisplay := websiteURL
	if parts := strings.Split(websiteURL, "//"); len(parts) > 1 {
		urlDisplay = parts[1]
	}

	message := descText + "Link: <a href=\"" + websiteURL + "\">" + urlDisplay + "</a>"

	c.SendAnswer(message)
}
