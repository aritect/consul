package buybot

import (
	"consul-telegram-bot/internal/bot"
	"consul-telegram-bot/internal/config"
	"consul-telegram-bot/internal/logger"
	"consul-telegram-bot/internal/model"
	"consul-telegram-bot/internal/utils"

	"fmt"
	"math/rand"
	"path/filepath"
	"time"

	telebot "gopkg.in/telebot.v3"
)

type SignalSender struct {
	bot      *bot.Bot
	logger   *logger.Logger
	config   *config.Config
	gifPaths []string
	rng      *rand.Rand
}

func NewSignalSender(botInstance *bot.Bot, logger *logger.Logger, cfg *config.Config) *SignalSender {
	gifPaths := []string{
		"./assets/1.gif",
		"./assets/2.gif",
		"./assets/3.gif",
		"./assets/4.gif",
		"./assets/5.gif",
		"./assets/6.gif",
	}

	return &SignalSender{
		bot:      botInstance,
		logger:   logger,
		config:   cfg,
		gifPaths: gifPaths,
		rng:      rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (s *SignalSender) SendBuySignal(buyTx *BuyTransaction) {
	s.logger.Info("sending buy signal for tx: %s", buyTx.Signature)

	recipients, err := s.getAllRecipients()
	if err != nil {
		s.logger.Error("failed to get recipients: %s", err)
		return
	}

	gifPath := s.getRandomGif()

	for _, recipient := range recipients {
		threadId := recipient.GetThreadIdForSignalType(model.SignalTypeBuys)

		if threadId == 0 {
			continue
		}

		ticker := model.GetWithFallback(recipient.TokenTicker, s.config.TokenTicker)
		dexURL := model.GetWithFallback(recipient.DexURL, s.config.DexURL)
		axiomURL := model.GetWithFallback(recipient.AxiomURL, s.config.AxiomURL)

		message := s.formatBuyMessage(buyTx, ticker)

		s.logger.Info("sending buy signal to chat %d, thread %d", recipient.Id, threadId)
		s.sendAnimationWithCaption(recipient, gifPath, message, threadId, dexURL, axiomURL)
	}
}

func (s *SignalSender) formatBuyMessage(buyTx *BuyTransaction, ticker string) string {
	if ticker == "" {
		ticker = "TOKEN"
	}

	return fmt.Sprintf(
		"<b>$%s BUY 🥬🥦🌿🌵🌳☘️</b>\n\n"+
			"<b>💰 Amount:</b> %s\n"+
			"<b>🦊 Buyer:</b> %s\n"+
			"<b>🔎 Transaction:</b> <a href=\"%s\">%s</a>",
		ticker,
		utils.FormatNumber(buyTx.Amount, ticker),
		s.shortenAddress(buyTx.Buyer),
		buyTx.TxURL,
		s.shortenAddress(buyTx.Signature),
	)
}

func (s *SignalSender) shortenAddress(address string) string {
	if len(address) <= 12 {
		return address
	}
	return address[:6] + "..." + address[len(address)-4:]
}

func (s *SignalSender) getAllRecipients() ([]*model.Recipient, error) {
	allRecipients := model.FindAllRecipients()

	var recipients []*model.Recipient
	for _, recipient := range allRecipients {
		if recipient.Receiving == 1 {
			recipients = append(recipients, recipient)
		}
	}

	return recipients, nil
}

func (s *SignalSender) getRandomGif() string {
	if len(s.gifPaths) == 0 {
		return ""
	}

	idx := s.rng.Intn(len(s.gifPaths))
	absPath, _ := filepath.Abs(s.gifPaths[idx])
	return absPath
}

func (s *SignalSender) sendAnimationWithCaption(recipient *model.Recipient, gifPath string, caption string, threadId int, dexURL string, axiomURL string) {
	animation := &telebot.Animation{
		File:     telebot.FromDisk(gifPath),
		MIME:     "image/gif",
		Caption:  caption,
		FileName: "animation.gif",
	}

	inlineKeyboard := &telebot.ReplyMarkup{}

	var buttons []telebot.Btn
	if dexURL != "" {
		buttons = append(buttons, inlineKeyboard.URL("Buy on Dexscreener", dexURL))
	}
	if axiomURL != "" {
		buttons = append(buttons, inlineKeyboard.URL("Buy on Axiom", axiomURL))
	}

	if len(buttons) > 0 {
		inlineKeyboard.Inline(inlineKeyboard.Row(buttons...))
	}

	opts := &telebot.SendOptions{
		ParseMode:   "HTML",
		ThreadID:    threadId,
		ReplyMarkup: inlineKeyboard,
	}

	_, err := s.bot.Bot.Send(recipient, animation, caption, opts)
	if err != nil {
		s.logger.Error("failed to send animation to chat %d: %s", recipient.Id, err)
	}
}
