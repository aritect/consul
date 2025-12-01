package commands

import (
	"consul-telegram-bot/internal/metrics"
	"consul-telegram-bot/internal/middlewares"
	"consul-telegram-bot/internal/model"
	"consul-telegram-bot/internal/router"
	"fmt"
)

func Setup(c *router.Context) {
	middlewares.Manager(setupHandler, c.Config.ManagerId)(c)
}

func setupHandler(c *router.Context) {
	chat := c.Message.Chat
	recipient, err := model.FindRecipient(chat.ID)
	if err != nil {
		metrics.TelegramCommandsProcessed.WithLabelValues("setup", "error").Inc()
		c.SendAnswer("🚧 Please run /start first to initialize the bot.")
		return
	}

	metrics.TelegramCommandsProcessed.WithLabelValues("setup", "success").Inc()

	configured := 0
	total := 7

	fields := []struct {
		name     string
		value    string
		fallback string
		desc     string
	}{
		{"name", recipient.ProjectName, c.Config.ProjectName, "Project name"},
		{"ticker", recipient.TokenTicker, c.Config.TokenTicker, "Token ticker"},
		{"description", recipient.Description, c.Config.Description, "Description"},
		{"website_url", recipient.WebsiteURL, c.Config.WebsiteURL, "Website URL"},
		{"token_address", recipient.TokenAddress, c.Config.TokenAddress, "Token address"},
		{"dex_url", recipient.DexURL, c.Config.DexURL, "Dexscreener URL"},
		{"axiom_url", recipient.AxiomURL, c.Config.AxiomURL, "Axiom URL"},
	}

	status := ""
	for _, f := range fields {
		icon := "⬜"
		source := ""
		if f.value != "" {
			icon = "✅"
			configured++
		} else if f.fallback != "" {
			icon = "⚙️"
			source = " (env)"
		}
		status += fmt.Sprintf("%s <b>%s</b>%s\n", icon, f.desc, source)
	}

	message := fmt.Sprintf(
		"<b>🔧 Setup Wizard</b>\n\n"+
			"<b>Progress:</b> %d/%d configured\n"+
			"✅ = set via /set\n"+
			"⚙️ = from env\n"+
			"👻 = not configured\n\n"+
			"%s\n"+
			"<b>Commands:</b>\n"+
			"<code>/set name Aritect</code>\n"+
			"<code>/set ticker TOKEN</code>\n"+
			"<code>/set description Your description</code>\n"+
			"<code>/set website_url https://example.com</code>\n"+
			"<code>/set token_address ABC123...</code>\n"+
			"<code>/set dex_url https://dexscreener.com/...</code>\n"+
			"<code>/set axiom_url https://axiom.trade/...</code>",
		configured,
		total,
		status,
	)

	c.SendAnswer(message)
}
