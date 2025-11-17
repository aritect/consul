package main

import (
	"consul-telegram-bot/internal/bot"
	"consul-telegram-bot/internal/commands"
	"consul-telegram-bot/internal/config"
	"consul-telegram-bot/internal/logger"
	"consul-telegram-bot/internal/metrics"
	"consul-telegram-bot/internal/router"
	"consul-telegram-bot/internal/store"

	telebot "gopkg.in/telebot.v3"
)

func startUpdatesListener(botInstance *bot.Bot, routerInstance *router.Router) {
	botInstance.Bot.Handle(telebot.OnText, func(c telebot.Context) error {
		routerInstance.HandleTextMessage(c.Message())
		return nil
	})
}

func configureKeyboard(botInstance *bot.Bot) {
	defaultKeyboard := [][]telebot.ReplyButton{
		{telebot.ReplyButton{Text: "Id"}, telebot.ReplyButton{Text: "Help"}},
	}

	botInstance.SetKeyboard(defaultKeyboard)
}

func configureCommands(routerInstance *router.Router) {
	routerInstance.AddCommand("/start", commands.Start)
	routerInstance.AddCommand("/id", commands.Id)
	routerInstance.AddCommand("/help", commands.Help)
	routerInstance.AddCommand("/website", commands.Website)
	routerInstance.AddCommand("/ca", commands.CA)
	routerInstance.AddCommand("/arbiter", commands.Arbiter)
	routerInstance.AddCommand("/agartha", commands.Agartha)
	routerInstance.AddCommand("/chart", commands.Chart)

	routerInstance.LinkingButton("Help", "/help")
	routerInstance.LinkingButton("Id", "/id")
}

func main() {
	loggerInstance := logger.New()

	go func() {
		loggerInstance.Info("starting metrics server on port 8080...")
		metrics.StartMetricsServer("8080")
	}()

	loggerInstance.Info("creating config instance...")
	configInstance := config.New()

	loggerInstance.Info("creating store instance...")
	storeInstance, err := store.New(configInstance.StorePath, false, false)
	if err != nil {
		panic(err)
	}

	storeInstance.MakeGlobal()

	loggerInstance.Info("creating bot instance...")
	botInstance, err := bot.New(configInstance.TelegramBotToken, 10)
	if err != nil {
		panic(err)
	}

	loggerInstance.Info("creating router instance...")
	routerInstance := router.New(botInstance, loggerInstance, configInstance)
	configureCommands(routerInstance)
	configureKeyboard(botInstance)

	go startUpdatesListener(botInstance, routerInstance)

	botInstance.Start(8)
}
