package commands

import (
	"consul-telegram-bot/internal/metrics"
	"consul-telegram-bot/internal/model"
	"consul-telegram-bot/internal/router"
	"fmt"
)

func Up(c *router.Context) {
	if c.Message.ReplyTo == nil {
		metrics.TelegramCommandsProcessed.WithLabelValues("up", "error").Inc()
		c.SendAnswer("⚠️ Reply to a message with /up to give a point to its author.")
		return
	}

	replyTo := c.Message.ReplyTo

	if replyTo.Sender == nil {
		metrics.TelegramCommandsProcessed.WithLabelValues("up", "error").Inc()
		c.SendAnswer("⚠️ Cannot identify the message author.")
		return
	}

	if replyTo.Sender.ID == c.Message.Sender.ID {
		metrics.TelegramCommandsProcessed.WithLabelValues("up", "error").Inc()
		c.SendAnswer("⚠️ You cannot give points to yourself.")
		return
	}

	if replyTo.Sender.IsBot {
		metrics.TelegramCommandsProcessed.WithLabelValues("up", "error").Inc()
		c.SendAnswer("⚠️ You cannot give points to bots.")
		return
	}

	chatID := c.Message.Chat.ID
	voterID := c.Message.Sender.ID
	targetID := replyTo.Sender.ID

	canVote, err := model.CanUserVote(chatID, voterID, targetID)
	if err != nil {
		metrics.TelegramCommandsProcessed.WithLabelValues("up", "error").Inc()
		metrics.ErrorsTotal.WithLabelValues("command", "up_check_vote").Inc()
		c.SendAnswer("🚧 Something went wrong. Please try again later.")
		return
	}

	if !canVote {
		metrics.TelegramCommandsProcessed.WithLabelValues("up", "cooldown").Inc()
		c.SendAnswer("⏳ You can give a point to this user again in an hour.")
		return
	}

	targetName := getDisplayName(replyTo.Sender.FirstName, replyTo.Sender.LastName, replyTo.Sender.Username)
	targetUsername := replyTo.Sender.Username

	rating, err := model.GetOrCreateUserRating(chatID, targetID, targetUsername, targetName)
	if err != nil {
		metrics.TelegramCommandsProcessed.WithLabelValues("up", "error").Inc()
		metrics.ErrorsTotal.WithLabelValues("command", "up_get_rating").Inc()
		c.SendAnswer("🚧 Something went wrong. Please try again later.")
		return
	}

	if err := rating.AddPoint(); err != nil {
		metrics.TelegramCommandsProcessed.WithLabelValues("up", "error").Inc()
		metrics.ErrorsTotal.WithLabelValues("command", "up_add_point").Inc()
		c.SendAnswer("🚧 Something went wrong. Please try again later.")
		return
	}

	if err := model.RecordVote(chatID, voterID, targetID); err != nil {
		metrics.TelegramCommandsProcessed.WithLabelValues("up", "error").Inc()
		metrics.ErrorsTotal.WithLabelValues("command", "up_record_vote").Inc()
	}

	metrics.TelegramCommandsProcessed.WithLabelValues("up", "success").Inc()

	voterName := getDisplayName(c.Message.Sender.FirstName, c.Message.Sender.LastName, c.Message.Sender.Username)
	response := fmt.Sprintf("⬆️ <b>%s</b> gave a point to <b>%s</b>!\n\n🏆 <b>%s</b> now has <b>%d</b> point(s).",
		voterName, targetName, targetName, rating.Points)

	c.SendAnswer(response)
}

func getDisplayName(firstName, lastName, username string) string {
	name := firstName
	if lastName != "" {
		if name != "" {
			name += " "
		}
		name += lastName
	}
	if name == "" && username != "" {
		name = username
	}
	if name == "" {
		name = "Anonymous"
	}
	return name
}
