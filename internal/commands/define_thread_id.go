package commands

import (
	"consul-telegram-bot/internal/metrics"
	"consul-telegram-bot/internal/model"
	"consul-telegram-bot/internal/router"
)

func DefineThreadId(c *router.Context) {
	recipient, err := model.FindRecipient(c.Message.Chat.ID)
	if err != nil {
		metrics.TelegramCommandsProcessed.WithLabelValues("define_thread_id", "error").Inc()
		c.SendAnswer("🚧 Recipient is not found.")
		return
	}

	if len(c.Args) == 0 {
		metrics.TelegramCommandsProcessed.WithLabelValues("define_thread_id", "error").Inc()
		c.SendAnswer("🚧 Please specify signal type: aritect_buys, retransmit.")
		return
	}

	signalTypeStr := c.Args[0]
	signalType, ok := model.ParseSignalType(signalTypeStr)
	if !ok {
		metrics.TelegramCommandsProcessed.WithLabelValues("define_thread_id", "error").Inc()
		c.SendAnswer("🚧 Invalid signal type. Available types: aritect_buys, retransmit.")
		return
	}

	recipient.DefineThreadIdForSignalType(signalType, c.Message.ThreadID)
	err = recipient.Write()
	if err != nil {
		metrics.TelegramCommandsProcessed.WithLabelValues("define_thread_id", "error").Inc()
		metrics.ErrorsTotal.WithLabelValues("command", "database_write").Inc()
		c.SendAnswer("🚧 Unfortunately, something went wrong.")
		return
	}

	metrics.TelegramCommandsProcessed.WithLabelValues("define_thread_id", "success").Inc()
	c.SendAnswer("🙆‍♀️ Thread ID has been defined for " + signalType.String() + ".")
}
