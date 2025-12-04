package router

import (
	"consul-telegram-bot/internal/bot"
	"consul-telegram-bot/internal/config"
	"consul-telegram-bot/internal/logger"
	"consul-telegram-bot/internal/metrics"
	"consul-telegram-bot/internal/model"
	"runtime"
	"strings"
	"sync"
	"time"

	telebot "gopkg.in/telebot.v3"
)

type Callback func(m *Context)

type Router struct {
	mu       sync.Mutex
	wg       sync.WaitGroup
	commands map[string]Callback
	bot      *bot.Bot
	config   *config.Config
	logger   *logger.Logger
	buttons  map[string]string
}

func New(b *bot.Bot, l *logger.Logger, c *config.Config) *Router {
	return &Router{
		commands: make(map[string]Callback),
		buttons:  make(map[string]string),
		bot:      b,
		logger:   l,
		config:   c,
	}
}

func (r *Router) AddCommand(command string, callback Callback) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.commands[command] = callback
}

func (r *Router) LinkingButton(button string, command string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.buttons[button] = command
}

func (r *Router) HandleTextMessage(m *telebot.Message) {
	now := time.Now()
	if now.Unix()-int64(m.Unixtime) > 60 {
		r.logger.Info("skipped message from %d", m.Chat.ID)
		return
	}

	r.logger.Info("received message from %d", m.Chat.ID)

	chatType := m.Chat.Type

	if !strings.HasPrefix(m.Text, "/") && (chatType == telebot.ChatGroup || chatType == telebot.ChatSuperGroup) {
		r.storeMessage(m)
	}

	command := "unknown"
	if strings.HasPrefix(m.Text, "/") {
		command = strings.Split(m.Text, " ")[0]
		r.logger.Info("received message with command %s", command)
		r.handleCommand(*m)
	} else {
		command = "button"
		r.handleCommandButton(*m)
	}

	if command == "button" {
		if button, ok := r.buttons[m.Text]; ok {
			metrics.TelegramMessagesReceived.WithLabelValues(string(chatType), button).Inc()
		} else {
			metrics.TelegramMessagesReceived.WithLabelValues(string(chatType), "unknown").Inc()
		}
	} else {
		metrics.TelegramMessagesReceived.WithLabelValues(string(chatType), command).Inc()
	}

	r.wg.Wait()

	elapsed := time.Since(now)
	metrics.ProcessingDuration.WithLabelValues("message_handling").Observe(float64(elapsed.Nanoseconds() / 1000))

	r.logger.Info("message processing from %d completed in %dµs", m.Chat.ID, elapsed.Nanoseconds()/1000)
}

func (r *Router) handleCommandButton(m telebot.Message) {
	execFn := r.commands[r.buttons[m.Text]]
	msg := r.parseMessage(&m)

	if execFn != nil {
		r.Safely(func() {
			execFn(msg)
		})
	}
}

func (r *Router) handleCommand(m telebot.Message) {
	msg := r.parseMessage(&m)
	execFn := r.commands[msg.Command]

	if execFn != nil {
		r.Safely(func() {
			execFn(msg)
		})
	}
}

func (r *Router) parseMessage(m *telebot.Message) *Context {
	command := ""
	var args []string

	if m.Text != "" {
		msgTokens := strings.Fields(m.Text)
		command, args = strings.ToLower(msgTokens[0]), msgTokens[1:]
		if strings.Contains(command, "@") {
			splittedCommand := strings.Split(command, "@")
			command = splittedCommand[0]
		}
	}

	return &Context{
		Args:    args,
		Command: command,
		Bot:     r.bot,
		Logger:  r.logger,
		Config:  r.config,
		Message: m,
	}
}

func (r *Router) Safely(fn func()) {
	r.wg.Add(1)

	go func() {
		defer func() {
			if err := recover(); err != nil {
				stack := make([]byte, 1024*8)
				stack = stack[:runtime.Stack(stack, false)]

				r.logger.Error("%s\n%s", err, stack)
			}

			r.wg.Done()
		}()

		fn()
	}()
}

func (r *Router) storeMessage(m *telebot.Message) {
	senderName := ""
	senderUsername := ""
	if m.Sender != nil {
		if m.Sender.FirstName != "" {
			senderName = m.Sender.FirstName
		}
		if m.Sender.LastName != "" {
			if senderName != "" {
				senderName += " "
			}
			senderName += m.Sender.LastName
		}
		if senderName == "" && m.Sender.Username != "" {
			senderName = m.Sender.Username
		}
		senderUsername = m.Sender.Username
	}

	senderID := int64(0)
	if m.Sender != nil {
		senderID = m.Sender.ID
	}

	_, err := model.NewMessage(
		m.Chat.ID,
		m.ID,
		senderID,
		senderName,
		senderUsername,
		m.Text,
		int64(m.Unixtime),
	)

	if err != nil {
		r.logger.Error("failed to store message: %s", err)
		return
	}

	go func() {
		deleted, err := model.DeleteExcessMessages(m.Chat.ID, 200)
		if err != nil {
			r.logger.Error("failed to cleanup old messages: %s", err)
		} else if deleted > 0 {
			r.logger.Info("cleaned up %d old messages from chat %d", deleted, m.Chat.ID)
		}
	}()
}
