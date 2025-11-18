package buybot

import (
	"consul-telegram-bot/internal/bot"
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
	bot            *bot.Bot
	logger         *logger.Logger
	gifPaths       []string
	rng            *rand.Rand
	dexscreenerUrl string
	axiomUrl       string
}

func NewSignalSender(botInstance *bot.Bot, logger *logger.Logger, dexscreenerUrl string, axiomUrl string) *SignalSender {
	gifPaths := []string{
		"./assets/1.gif",
		"./assets/2.gif",
		"./assets/3.gif",
		"./assets/4.gif",
		"./assets/5.gif",
		"./assets/6.gif",
	}

	return &SignalSender{
		bot:            botInstance,
		logger:         logger,
		gifPaths:       gifPaths,
		rng:            rand.New(rand.NewSource(time.Now().UnixNano())),
		dexscreenerUrl: dexscreenerUrl,
		axiomUrl:       axiomUrl,
	}
}

func (s *SignalSender) SendBuySignal(buyTx *BuyTransaction) {
	s.logger.Info("sending buy signal for tx: %s", buyTx.Signature)

	recipients, err := s.getAllRecipients()
	if err != nil {
		s.logger.Error("failed to get recipients: %s", err)
		return
	}

	message := s.formatBuyMessage(buyTx)
	gifPath := s.getRandomGif()

	for _, recipient := range recipients {
		threadId := recipient.GetThreadIdForSignalType(model.SignalTypeAritectBuys)

		if threadId == 0 {
			continue
		}

		s.logger.Info("sending buy signal to chat %d, thread %d", recipient.Id, threadId)
		s.sendAnimationWithCaption(recipient, gifPath, message, threadId)
	}
}

func (s *SignalSender) formatBuyMessage(buyTx *BuyTransaction) string {
	if buyTx.SolAmount == 0 {
		return fmt.Sprintf(
			"<b>$ARITECT BUY 🥬🥦🌿🌵🌳☘️</b>\n\n"+
				"<b>💰 Amount:</b> %s\n"+
				"<b>🦊 Buyer:</b> %s\n"+
				"<b>🔎 Transaction:</b> <a href=\"%s\">%s</a>",
			utils.FormatNumber(buyTx.Amount, "ARITECT"),
			s.shortenAddress(buyTx.Buyer),
			buyTx.TxURL,
			s.shortenAddress(buyTx.Signature),
		)
	}

	message := fmt.Sprintf(
		"<b>$ARITECT BUY 🥬🥦🌿🌵🌳☘️</b>\n\n"+
			"<b>💰 Amount:</b> %s\n"+
			"<b>🪙 SOL Spent:</b> %.4f SOL\n"+
			"<b>🦊 Buyer:</b> %s\n"+
			"<b>🔎 Transaction:</b> <a href=\"%s\">%s</a>",
		utils.FormatNumber(buyTx.Amount, "ARITECT"),
		buyTx.SolAmount,
		s.shortenAddress(buyTx.Buyer),
		buyTx.TxURL,
		s.shortenAddress(buyTx.Signature),
	)

	return message
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

func (s *SignalSender) sendAnimationWithCaption(recipient *model.Recipient, gifPath string, caption string, threadId int) {
	animation := &telebot.Animation{
		File:     telebot.FromDisk(gifPath),
		MIME:     "image/gif",
		Caption:  caption,
		FileName: "animation.gif",
	}

	inlineKeyboard := &telebot.ReplyMarkup{}
	dexScreenerBuyButton := inlineKeyboard.URL("Buy on Dexscreener", s.dexscreenerUrl)
	axiomBuyButton := inlineKeyboard.URL("Buy on Axiom", s.axiomUrl)

	inlineKeyboard.Inline(
		inlineKeyboard.Row(dexScreenerBuyButton, axiomBuyButton),
	)

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
