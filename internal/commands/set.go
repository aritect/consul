package commands

import (
	"consul-telegram-bot/internal/metrics"
	"consul-telegram-bot/internal/middlewares"
	"consul-telegram-bot/internal/model"
	"consul-telegram-bot/internal/router"
	"strings"
)

func Set(c *router.Context) {
	middlewares.Manager(setHandler, c.Config.ManagerId)(c)
}

func setHandler(c *router.Context) {
	if len(c.Args) < 2 {
		metrics.TelegramCommandsProcessed.WithLabelValues("set", "error").Inc()
		c.SendAnswer(
			"🚧 Usage: /set <field> <value>\n\n" +
				"<b>Fields:</b>\n" +
				"<code>name</code> — Project name\n" +
				"<code>ticker</code> — Token ticker\n" +
				"<code>description</code> — Description\n" +
				"<code>website_url</code> — Website URL\n" +
				"<code>token_address</code> — Token address\n" +
				"<code>dex_url</code> — Dexscreener URL\n" +
				"<code>axiom_url</code> — Axiom URL",
		)
		return
	}

	chat := c.Message.Chat
	recipient, err := model.FindRecipient(chat.ID)
	if err != nil {
		metrics.TelegramCommandsProcessed.WithLabelValues("set", "error").Inc()
		c.SendAnswer("🚧 Please run /start first.")
		return
	}

	field := strings.ToLower(c.Args[0])
	value := strings.Join(c.Args[1:], " ")

	var fieldName string
	switch field {
	case "name":
		recipient.ProjectName = value
		fieldName = "Project name"
	case "ticker":
		recipient.TokenTicker = strings.ToUpper(value)
		fieldName = "Token ticker"
	case "description":
		recipient.Description = value
		fieldName = "Description"
	case "website_url", "website":
		recipient.WebsiteURL = value
		fieldName = "Website URL"
	case "token_address", "ca", "address":
		recipient.TokenAddress = value
		fieldName = "Token address"
	case "dex_url", "dex":
		recipient.DexURL = value
		fieldName = "Dexscreener URL"
	case "axiom_url", "axiom":
		recipient.AxiomURL = value
		fieldName = "Axiom URL"
	default:
		metrics.TelegramCommandsProcessed.WithLabelValues("set", "error").Inc()
		c.SendAnswer("🚧 Unknown field: " + field)
		return
	}

	err = recipient.Write()
	if err != nil {
		metrics.TelegramCommandsProcessed.WithLabelValues("set", "error").Inc()
		c.SendAnswer("🚧 Failed to save.")
		return
	}

	metrics.TelegramCommandsProcessed.WithLabelValues("set", "success").Inc()
	c.SendAnswer("✅ " + fieldName + " updated.")
}
