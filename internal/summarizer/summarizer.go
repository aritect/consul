package summarizer

import (
	"consul-telegram-bot/internal/llm"
	"consul-telegram-bot/internal/model"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

type Summarizer struct {
	client *llm.Client
}

func New(client *llm.Client) *Summarizer {
	return &Summarizer{
		client: client,
	}
}

func (s *Summarizer) GenerateSummary(messages []*model.Message, projectName string) (string, error) {
	if len(messages) == 0 {
		return "", fmt.Errorf("no messages to summarize")
	}

	chatHistory := s.formatMessages(messages)
	prompt := s.buildPrompt(chatHistory, projectName, len(messages))

	response, err := s.client.ChatWithOptions(
		[]llm.ChatMessage{
			{
				Role:    "system",
				Content: s.getSystemPrompt(),
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
		1024,
		0.5,
	)

	if err != nil {
		return "", fmt.Errorf("failed to generate summary: %w", err)
	}

	result := s.replaceIndexesWithLinks(response, messages)
	result = s.replaceParticipantsPlaceholder(result, messages)
	result = s.normalizeEmptyLines(result)
	return result, nil
}

func (s *Summarizer) formatMessages(messages []*model.Message) string {
	var sb strings.Builder

	for i, msg := range messages {
		name := msg.SenderName
		if name == "" {
			name = "Anonymous"
		}
		sb.WriteString(fmt.Sprintf("[#%d][%s]: %s\n", i+1, name, msg.Text))
	}

	return sb.String()
}

func (s *Summarizer) replaceIndexesWithLinks(text string, messages []*model.Message) string {
	if len(messages) == 0 {
		return text
	}

	chatID := messages[0].ChatID
	re := regexp.MustCompile(`\(#(\d+),\s*(\d+\s*\S+)\)`)

	return re.ReplaceAllStringFunc(text, func(match string) string {
		submatches := re.FindStringSubmatch(match)
		if len(submatches) < 3 {
			return match
		}

		index, err := strconv.Atoi(submatches[1])
		if err != nil || index < 1 || index > len(messages) {
			return match
		}

		msg := messages[index-1]
		link := s.buildMessageLink(chatID, msg.MessageID)
		messageCount := submatches[2]
		return fmt.Sprintf("(<a href=\"%s\">%s</a>)", link, messageCount)
	})
}

func (s *Summarizer) buildMessageLink(chatID int64, messageID int) string {
	if chatID < 0 {
		chatIDStr := strconv.FormatInt(-chatID, 10)
		if len(chatIDStr) > 3 && chatIDStr[:3] == "100" {
			chatIDStr = chatIDStr[3:]
		}
		return fmt.Sprintf("https://t.me/c/%s/%d", chatIDStr, messageID)
	}
	return fmt.Sprintf("https://t.me/c/%d/%d", chatID, messageID)
}

func (s *Summarizer) normalizeEmptyLines(text string) string {
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimRight(line, " \t")
	}
	text = strings.Join(lines, "\n")

	re := regexp.MustCompile(`\n{3,}`)
	return re.ReplaceAllString(text, "\n\n")
}

func (s *Summarizer) replaceParticipantsPlaceholder(text string, messages []*model.Message) string {
	usernameCounts := make(map[string]int)
	for _, msg := range messages {
		if msg.SenderUsername != "" {
			usernameCounts[msg.SenderUsername]++
		}
	}

	if len(usernameCounts) == 0 {
		return strings.Replace(text, "{{PARTICIPANTS_PLACEHOLDER}}", "", 1)
	}

	type participant struct {
		username string
		count    int
	}

	participants := make([]participant, 0, len(usernameCounts))
	for username, count := range usernameCounts {
		participants = append(participants, participant{username, count})
	}

	sort.Slice(participants, func(i, j int) bool {
		return participants[i].count > participants[j].count
	})

	maxShow := 3
	if len(participants) < maxShow {
		maxShow = len(participants)
	}

	var mentions []string
	for i := 0; i < maxShow; i++ {
		mentions = append(mentions, "@"+participants[i].username)
	}

	participantsLine := "💬 Active voices: " + strings.Join(mentions, ", ")
	if len(participants) > maxShow {
		participantsLine += " and others"
	}
	participantsLine += "."

	return strings.Replace(text, "{{PARTICIPANTS_PLACEHOLDER}}", participantsLine, 1)
}

func (s *Summarizer) getSystemPrompt() string {
	return `You are a helpful assistant that summarizes community chat conversations.
Your task is to create a structured summary of the chat messages.

Output format (STRICTLY follow this structure):
1. First line: Start with "💁🏼‍♀️" emoji, then a brief overall summary (maximum 300 characters).
2. ONE empty line (not two).
3. List of discussion topics. Each line starts with "- " and format: "- Topic description (#N, M messages)."
   - #N is the message number where this topic FIRST appears (will become a link)
   - M is the number of messages related to this topic
4. ONE empty line.
5. Output the exact text: {{PARTICIPANTS_PLACEHOLDER}} (do NOT modify, translate, or add anything around this token).
6. ONE empty line.
7. Final line: A short motivational thought, wish, or friendly closing remark related to the community.

Example output:
💁🏼‍♀️ Community discussed technical issues and shared resources for the project.

- Screen replacement and display problems (#3, 5 messages).
- Equipment spare parts search (#12, 3 messages).
- Project updates and announcements (#18, 2 messages).

{{PARTICIPANTS_PLACEHOLDER}}

Keep building great things together! 🚀

Guidelines:
- Write in the same language as the majority of messages.
- Group related messages into topics.
- Reference the FIRST message of each topic using #N format.
- Ignore spam, greetings, and off-topic messages.
- Keep topic descriptions concise (under 60 characters each).
- List 3-7 most significant topics.
- The closing thought should be warm, encouraging, and relevant.

IMPORTANT: Use plain text only. No markdown formatting (**, *, bullet points except for the list). Use exactly ONE empty line between sections.`
}

func (s *Summarizer) buildPrompt(chatHistory, projectName string, messageCount int) string {
	projectContext := ""
	if projectName != "" {
		projectContext = fmt.Sprintf(" for the %s community", projectName)
	}

	return fmt.Sprintf(`Please summarize the following %d chat messages%s.
Provide a clear and structured summary that helps community members catch up on what they missed.

Chat messages:
%s

Please provide a summary:`, messageCount, projectContext, chatHistory)
}
