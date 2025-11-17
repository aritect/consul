package router

import (
	"consul-telegram-bot/internal/bot"
	"consul-telegram-bot/internal/config"
	"consul-telegram-bot/internal/logger"
	"consul-telegram-bot/internal/model"
	"strings"

	telebot "gopkg.in/telebot.v3"
)

type Context struct {
	Args    []string
	Command string
	Message *telebot.Message
	Bot     *bot.Bot
	Logger  *logger.Logger
	Config  *config.Config
}

func (c Context) GetArgString() string {
	argString := ""
	for _, s := range c.Args {
		argString += s + " "
	}
	return strings.TrimSpace(argString)
}

func (c *Context) IsManager() bool {
	return c.Message.Sender.ID == c.Config.ManagerId
}

func (c *Context) SendAnswer(text string) {
	recipient, err := model.FindRecipient(c.Message.Chat.ID)
	if err != nil {
		c.Logger.Error("error finding recipient: %s", err)
		return
	}

	c.Bot.SendWithLimit(recipient, text, false, true, c.Message.ThreadID, false)
}
