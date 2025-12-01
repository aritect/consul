package commands

import (
	"consul-telegram-bot/internal/metrics"
	"consul-telegram-bot/internal/middlewares"
	"consul-telegram-bot/internal/model"
	"consul-telegram-bot/internal/router"
	"fmt"
	"log"
)

func Retransmit(c *router.Context) {
	middlewares.Manager(retransmitHandler, c.Config.ManagerId)(c)
}

func retransmitHandler(c *router.Context) {
	if len(c.Args) == 0 {
		metrics.TelegramCommandsProcessed.WithLabelValues("retransmit", "error").Inc()
		c.SendAnswer("🚧 Please provide a message to retransmit.")
		return
	}

	message := c.GetArgString()

	recipients := model.FindAllRecipients()
	sentCount := 0

	for _, recipient := range recipients {
		threadId := recipient.GetThreadIdForSignalType(model.SignalTypeRetransmit)
		log.Println(threadId)
		if threadId == 0 {
			continue
		}

		c.Bot.SendWithLimit(recipient, message, false, false, threadId, false)
		sentCount++
	}

	metrics.TelegramCommandsProcessed.WithLabelValues("retransmit", "success").Inc()
	c.SendAnswer(fmt.Sprintf("✅ Message retransmitted to %d recipients.", sentCount))
}
