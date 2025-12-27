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

func (c Context) GetArgStringWithNewlines() string {
	if c.Message == nil {
		return ""
	}

	text := c.Message.Text
	if text == "" && c.Message.Caption != "" {
		text = c.Message.Caption
	}

	if text == "" {
		return ""
	}

	commandEnd := strings.Index(text, "\n")

	if commandEnd == -1 {
		parts := strings.SplitN(text, " ", 2)
		if len(parts) < 2 {
			return ""
		}
		return parts[1]
	}

	return strings.TrimPrefix(text[commandEnd:], "\n")
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
