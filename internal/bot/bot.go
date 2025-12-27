package bot

import (
	"consul-telegram-bot/internal/metrics"
	"consul-telegram-bot/internal/model"
	"errors"
	"time"

	"github.com/soluchok/tsender"
	telebot "gopkg.in/telebot.v3"
)

type Bot struct {
	Bot     *telebot.Bot
	sender  *tsender.Sender
	options *telebot.SendOptions
}

type OutputMessage struct {
	Text                        string
	NeedReply                   bool
	SameThread                  bool
	ThreadId                    int
	Recipient                   *model.Recipient
	WithoutNotificationForGroup bool
	ReplyToMessageID            int
}

func New(token string, pollingTimeout int) (*Bot, error) {
	settings := telebot.Settings{
		Token:   token,
		Updates: 100,
		Poller: &telebot.LongPoller{
			Timeout: time.Duration(pollingTimeout) * time.Second,
		},
	}

	tb, err := telebot.NewBot(settings)
	if err != nil {
		return nil, err
	}

	bot := &Bot{
		Bot: tb,
	}

	replyMarkup := telebot.ReplyMarkup{
		Selective:       false,
		ForceReply:      false,
		ResizeKeyboard:  true,
		ReplyKeyboard:   [][]telebot.ReplyButton{},
		OneTimeKeyboard: true,
	}

	bot.options = &telebot.SendOptions{
		ParseMode:             "HTML",
		ReplyMarkup:           &replyMarkup,
		DisableWebPagePreview: false,
	}

	return bot, nil
}

func (b *Bot) Start(workers int) {
	b.sender = tsender.NewSender(b)
	go b.sender.Run(workers)
	defer b.sender.Stop()

	b.Bot.Start()
}

func (b *Bot) SetKeyboard(keyboard [][]telebot.ReplyButton) {
	b.options.ReplyMarkup.OneTimeKeyboard = false
	b.options.ReplyMarkup.ReplyKeyboard = keyboard
}

func (b *Bot) Send(message interface{}) {
	defer b.reset()

	m, ok := message.(OutputMessage)

	if ok {
		if m.Recipient.Type == model.RecipientPrivate {
			b.options.ReplyMarkup.Selective = m.NeedReply
			b.options.ReplyMarkup.ForceReply = m.NeedReply
		} else {
			b.options.ReplyMarkup.Selective = false
			b.options.ReplyMarkup.ForceReply = false
			b.options.ReplyMarkup.ReplyKeyboard = nil
			b.options.ReplyMarkup.OneTimeKeyboard = false

			if m.WithoutNotificationForGroup {
				b.options.DisableNotification = true
			}
		}

		if m.ThreadId != 0 {
			b.options.ThreadID = m.ThreadId
		} else {
			b.options.ThreadID = m.Recipient.ThreadId
		}

		if m.ReplyToMessageID != 0 {
			b.options.ReplyTo = &telebot.Message{ID: m.ReplyToMessageID}
		}

		_, err := b.Bot.Send(m.Recipient, m.Text, b.options)
		if err != nil {

			tbErr := new(telebot.Error)
			errors.As(err, &tbErr)

			if tbErr.Code == 403 {
				m.Recipient.DeleteSelf()
			}

			metrics.TelegramMessagesSent.WithLabelValues(string(m.Recipient.Type), "error").Inc()
			metrics.ErrorsTotal.WithLabelValues("telegram_bot", "send_message").Inc()
			return
		}

		metrics.TelegramMessagesSent.WithLabelValues(string(m.Recipient.Type), "success").Inc()
	}
}

func (b *Bot) SendWithLimit(recipient *model.Recipient, text string, needReply bool, sameThread bool, threadId int, withoutNotionficationForGroup bool) {
	b.sender.Send(recipient.Id, OutputMessage{
		Text:                        text,
		ThreadId:                    threadId,
		SameThread:                  sameThread,
		NeedReply:                   needReply,
		Recipient:                   recipient,
		WithoutNotificationForGroup: withoutNotionficationForGroup,
	})
}

func (b *Bot) SendWithLimitAndReply(recipient *model.Recipient, text string, needReply bool, sameThread bool, threadId int, withoutNotionficationForGroup bool, replyToMessageID int) {
	b.sender.Send(recipient.Id, OutputMessage{
		Text:                        text,
		ThreadId:                    threadId,
		SameThread:                  sameThread,
		NeedReply:                   needReply,
		Recipient:                   recipient,
		WithoutNotificationForGroup: withoutNotionficationForGroup,
		ReplyToMessageID:            replyToMessageID,
	})
}

func (b *Bot) reset() {
	b.options.ReplyMarkup.Selective = false
	b.options.ReplyMarkup.ForceReply = false
	b.options.ReplyTo = nil
}
