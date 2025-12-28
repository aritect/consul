package commands

import (
	"consul-telegram-bot/internal/llm"
	"consul-telegram-bot/internal/metrics"
	"consul-telegram-bot/internal/model"
	"consul-telegram-bot/internal/router"
	"fmt"
	"sync"
	"time"
)

const (
	consulCooldown           = 5 * time.Second
	maxTokens                = 2048
	temperature              = 0.7
	recentMessagesLimit      = 10
	messagesAroundReplyCount = 5
)

var (
	consulMu       sync.Mutex
	lastConsulTime = make(map[int64]time.Time)
)

func Consul(c *router.Context) {
	chatID := c.Message.Chat.ID

	consulMu.Lock()
	if lastTime, exists := lastConsulTime[chatID]; exists {
		if time.Since(lastTime) < consulCooldown {
			remaining := consulCooldown - time.Since(lastTime)
			consulMu.Unlock()
			c.SendAnswer(fmt.Sprintf("⏳ Please wait %d seconds before asking again.", int(remaining.Seconds())))
			return
		}
	}
	lastConsulTime[chatID] = time.Now()
	consulMu.Unlock()

	if c.Config.LLMAPIKey == "" {
		metrics.TelegramCommandsProcessed.WithLabelValues("consul", "error").Inc()
		c.SendAnswer("🚧 Consul feature is not configured. Please set LLM_API_KEY.")
		return
	}

	userQuestion := ""
	replyToMessageID := 0
	var conversationContext []*model.Message
	isReplyToBot := false

	if c.Message.ReplyTo != nil {
		if c.Message.ReplyTo.Sender != nil && c.Message.ReplyTo.Sender.IsBot {
			isReplyToBot = true
			replyToMessageID = c.Message.ID
		} else {
			replyToMessageID = c.Message.ReplyTo.ID
		}

		userQuestion = c.GetArgStringWithNewlines()
		if userQuestion == "" {
			if isReplyToBot {
				userQuestion = c.Message.Text
			} else {
				userQuestion = "What do you think about this?"
			}
		}

		contextMessages, err := model.GetMessagesAroundTarget(chatID, c.Message.ReplyTo.ID, messagesAroundReplyCount, messagesAroundReplyCount)
		if err == nil && len(contextMessages) > 0 {
			conversationContext = contextMessages
		}
	} else {
		if len(c.Args) == 0 {
			metrics.TelegramCommandsProcessed.WithLabelValues("consul", "error").Inc()
			c.SendAnswer("🚧 Please provide a question.\n\nExample:\n/consul What role can utility token play in Aritect platform?")
			return
		}
		userQuestion = c.GetArgStringWithNewlines()

		recentMessages, err := model.GetRecentMessagesForContext(chatID, recentMessagesLimit)
		if err == nil && len(recentMessages) > 0 {
			conversationContext = recentMessages
		}
	}

	llmContext, err := model.FindLLMContext(chatID)
	customContext := ""
	if err == nil && llmContext != nil {
		customContext = llmContext.Context
	}

	provider, ok := llm.ParseProvider(c.Config.LLMProvider)
	if !ok {
		provider = llm.ProviderGroq
	}

	client := llm.NewClient(provider, c.Config.LLMAPIKey, c.Config.LLMModel)

	systemPrompt := buildSystemPrompt(customContext)

	var targetMessageText string
	if c.Message.ReplyTo != nil && !isReplyToBot {
		targetMessageText = c.Message.ReplyTo.Text
	}

	userPrompt := buildUserPrompt(userQuestion, conversationContext, targetMessageText, isReplyToBot)

	messages := []llm.ChatMessage{
		{
			Role:    "system",
			Content: systemPrompt,
		},
		{
			Role:    "user",
			Content: userPrompt,
		},
	}

	response, err := client.ChatWithOptions(messages, maxTokens, temperature)
	if err != nil {
		c.Logger.Error("failed to get LLM response: %s", err)
		metrics.TelegramCommandsProcessed.WithLabelValues("consul", "error").Inc()
		c.SendAnswer("🚧 Failed to get response. Please try again later.")
		return
	}

	recipient, err := model.FindRecipient(chatID)
	if err != nil {
		c.Logger.Error("error finding recipient: %s", err)
		metrics.TelegramCommandsProcessed.WithLabelValues("consul", "error").Inc()
		return
	}

	if replyToMessageID != 0 {
		c.Bot.SendWithLimitAndReply(recipient, response, false, true, c.Message.ThreadID, false, replyToMessageID)
	} else {
		c.Bot.SendWithLimit(recipient, response, false, true, c.Message.ThreadID, false)
	}

	metrics.TelegramCommandsProcessed.WithLabelValues("consul", "success").Inc()
}

func buildSystemPrompt(customContext string) string {
	basePrompt := `You are Consul, an intelligent AI assistant for a cryptocurrency community. Your role is to provide helpful, accurate, and engaging responses to community members' questions.

Key guidelines:
- Be professional, friendly, and concise.
- Provide accurate information based on the context provided.
- If you don't know something, admit it honestly.
- Use emojis sparingly and appropriately.
- Keep responses focused and relevant.
- Avoid making financial advice or price predictions.
- When analyzing conversation context, focus on key points and main topics discussed.
- Write in the same language as the question.

CRITICAL FORMATTING RULES (MUST FOLLOW):
- Use plain text only. NO markdown formatting whatsoever.
- NEVER use **, *, ~~, ` + "`" + `, or any other markdown syntax in your responses.
- Even if the context or examples contain markdown, DO NOT use it in your output.
- When creating lists, use "- " (dash with space) at the start of each line.
- Always end list items with a period.
- Use exactly ONE empty line between sections.
- Keep paragraphs concise and well-structured.

Example list format:
- Lending protocols.
- Centralized exchanges.
- Payment processors.
- Token issuers.

IMPORTANT: The context provided may contain markdown formatting. You must read and understand it, but NEVER reproduce markdown syntax in your responses. Always output plain text only.`

	if customContext != "" {
		return fmt.Sprintf("%s\n\nAdditional context:\n%s", basePrompt, customContext)
	}

	return basePrompt
}

func buildUserPrompt(question string, conversationContext []*model.Message, targetMessageText string, isReplyToBot bool) string {
	var prompt string

	if targetMessageText != "" {
		prompt = fmt.Sprintf("Target message to analyze:\n\"%s\"\n\n", targetMessageText)
		prompt += fmt.Sprintf("User question about this message: %s\n\n", question)

		if len(conversationContext) > 0 {
			prompt += "Additional conversation context (±5 messages around the target):\n\n"
			for i, msg := range conversationContext {
				name := msg.SenderName
				if name == "" {
					name = "Anonymous"
				}
				if msg.SenderUsername != "" {
					name = "@" + msg.SenderUsername
				}
				prompt += fmt.Sprintf("[%d] %s: %s\n", i+1, name, truncateText(msg.Text, 200))
			}
		}

		prompt += "\nIMPORTANT: Answer the user's question specifically about the target message. Use the additional context only if relevant."
		return prompt
	}

	if len(conversationContext) == 0 {
		return question
	}

	if isReplyToBot {
		prompt = "Previous conversation history (±5 messages around your last response):\n\n"
	} else {
		prompt = "Recent conversation context (last 10 messages):\n\n"
	}

	for i, msg := range conversationContext {
		name := msg.SenderName
		if name == "" {
			name = "Anonymous"
		}
		if msg.SenderUsername != "" {
			name = "@" + msg.SenderUsername
		}
		prompt += fmt.Sprintf("[%d] %s: %s\n", i+1, name, truncateText(msg.Text, 200))
	}

	prompt += fmt.Sprintf("\nQuestion: %s", question)
	return prompt
}

func truncateText(text string, maxLen int) string {
	if len(text) <= maxLen {
		return text
	}
	return text[:maxLen] + "..."
}
