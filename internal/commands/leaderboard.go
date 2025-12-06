package commands

import (
	"consul-telegram-bot/internal/metrics"
	"consul-telegram-bot/internal/model"
	"consul-telegram-bot/internal/router"
	"fmt"
	"strings"
)

func Leaderboard(c *router.Context) {
	chatID := c.Message.Chat.ID

	ratings, err := model.GetTopRatings(chatID, 10)
	if err != nil {
		metrics.TelegramCommandsProcessed.WithLabelValues("leaderboard", "error").Inc()
		metrics.ErrorsTotal.WithLabelValues("command", "leaderboard_get_ratings").Inc()
		c.SendAnswer("🚧 Something went wrong. Please try again later.")
		return
	}

	if len(ratings) == 0 {
		metrics.TelegramCommandsProcessed.WithLabelValues("leaderboard", "success").Inc()
		c.SendAnswer("🏆 <b>Community Leaderboard</b>\n\nNo ratings yet. Use /up to give points to helpful community members!")
		return
	}

	var sb strings.Builder
	sb.WriteString("🏆 <b>Community Leaderboard</b>\n\n")

	medals := []string{"🥇", "🥈", "🥉"}

	for i, rating := range ratings {
		var position string
		if i < 3 {
			position = medals[i]
		} else {
			position = fmt.Sprintf("%d.", i+1)
		}

		if rating.Username != "" {
			sb.WriteString(fmt.Sprintf("%s <b>@%s</b> — %d pts.\n", position, rating.Username, rating.Points))
			continue
		}

		displayName := rating.DisplayName
		if displayName == "" {
			displayName = "Anonymous"
		}

		sb.WriteString(fmt.Sprintf("%s <b>%s</b> — %d pts.\n", position, displayName, rating.Points))
	}

	sb.WriteString("\n<i>Reply to a message with /up to give points!</i>")

	metrics.TelegramCommandsProcessed.WithLabelValues("leaderboard", "success").Inc()
	c.SendAnswer(sb.String())
}
