package commands

import (
	"consul-telegram-bot/internal/metrics"
	"consul-telegram-bot/internal/middlewares"
	"consul-telegram-bot/internal/model"
	"consul-telegram-bot/internal/router"
)

func Clear(c *router.Context) {
	middlewares.Manager(clearHandler, c.Config.ManagerId)(c)
}

func clearHandler(c *router.Context) {
	recipient, err := model.FindRecipient(c.Message.Chat.ID)
	if err != nil {
		metrics.TelegramCommandsProcessed.WithLabelValues("clear", "error").Inc()
		c.SendAnswer("🚧 Recipient is not found.")
		return
	}

	recipient.ProjectName = ""
	recipient.TokenTicker = ""
	recipient.Description = ""
	recipient.WebsiteURL = ""
	recipient.TokenAddress = ""
	recipient.DexURL = ""
	recipient.AxiomURL = ""
	recipient.ThreadId = 0
	recipient.BuysThreadId = 0
	recipient.RetransmitThreadId = 0

	err = recipient.Write()
	if err != nil {
		metrics.TelegramCommandsProcessed.WithLabelValues("clear", "error").Inc()
		c.SendAnswer("🚧 Failed to clear settings.")
		return
	}

	metrics.TelegramCommandsProcessed.WithLabelValues("clear", "success").Inc()
	c.SendAnswer("✅ All settings have been cleared for this chat.")
}
