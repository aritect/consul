package middlewares

import (
	"consul-telegram-bot/internal/router"
)

func Manager(next router.Callback, managerId int64) router.Callback {
	return func(c *router.Context) {
		if c.Message.Sender.ID != managerId {
			c.SendAnswer("🙅‍ You can't use this command.")
			return
		}

		next(c)
	}
}
