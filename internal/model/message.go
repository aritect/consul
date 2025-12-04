package model

import (
	"bytes"
	"consul-telegram-bot/internal/store"
	"fmt"
	"sort"
	"time"

	"github.com/vmihailenco/msgpack/v5"
)

type Message struct {
	ChatID         int64  `msgpack:"chat_id"`
	MessageID      int    `msgpack:"message_id"`
	SenderID       int64  `msgpack:"sender_id"`
	SenderName     string `msgpack:"sender_name"`
	SenderUsername string `msgpack:"sender_username"`
	Text           string `msgpack:"text"`
	Timestamp      int64  `msgpack:"timestamp"`
}

func NewMessage(chatID int64, messageID int, senderID int64, senderName, senderUsername, text string, timestamp int64) (*Message, error) {
	m := &Message{
		ChatID:         chatID,
		MessageID:      messageID,
		SenderID:       senderID,
		SenderName:     senderName,
		SenderUsername: senderUsername,
		Text:           text,
		Timestamp:      timestamp,
	}

	if err := m.Save(); err != nil {
		return nil, err
	}

	return m, nil
}

func (m *Message) Save() error {
	key := GetMessageKey(m.ChatID, m.Timestamp, m.MessageID)
	data, err := msgpack.Marshal(m)
	if err != nil {
		return err
	}

	storeInstance := store.GetInstance()
	return storeInstance.Put(key, data)
}

func GetLastMessages(chatID int64, limit int) ([]*Message, error) {
	storeInstance := store.GetInstance()
	iterator := storeInstance.Iterator()
	defer iterator.Release()

	prefix := []byte(fmt.Sprintf("message:%d:", chatID))
	messages := make([]*Message, 0, limit)

	for iterator.Next() {
		key := iterator.Key()
		if !bytes.HasPrefix(key, prefix) {
			continue
		}

		var msg Message
		if err := msgpack.Unmarshal(iterator.Value(), &msg); err != nil {
			continue
		}
		messages = append(messages, &msg)
	}

	sort.Slice(messages, func(i, j int) bool {
		return messages[i].Timestamp > messages[j].Timestamp
	})

	if len(messages) > limit {
		messages = messages[:limit]
	}

	return messages, nil
}

func GetMessagesForSummary(chatID int64, limit int) ([]*Message, error) {
	messages, err := GetLastMessages(chatID, limit)
	if err != nil {
		return nil, err
	}

	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, nil
}

func DeleteOldMessages(chatID int64, maxAge time.Duration) (int, error) {
	storeInstance := store.GetInstance()
	iterator := storeInstance.Iterator()
	defer iterator.Release()

	prefix := []byte(fmt.Sprintf("message:%d:", chatID))
	cutoff := time.Now().Add(-maxAge).Unix()
	deleted := 0

	keysToDelete := make([][]byte, 0)

	for iterator.Next() {
		key := iterator.Key()
		if !bytes.HasPrefix(key, prefix) {
			continue
		}

		var msg Message
		if err := msgpack.Unmarshal(iterator.Value(), &msg); err != nil {
			continue
		}

		if msg.Timestamp < cutoff {
			keyCopy := make([]byte, len(key))
			copy(keyCopy, key)
			keysToDelete = append(keysToDelete, keyCopy)
		}
	}

	for _, key := range keysToDelete {
		if err := storeInstance.Delete(key); err == nil {
			deleted++
		}
	}

	return deleted, nil
}

func DeleteExcessMessages(chatID int64, keepCount int) (int, error) {
	messages, err := GetLastMessages(chatID, keepCount*2)
	if err != nil {
		return 0, err
	}

	if len(messages) <= keepCount {
		return 0, nil
	}

	storeInstance := store.GetInstance()
	deleted := 0

	for i := keepCount; i < len(messages); i++ {
		key := GetMessageKey(messages[i].ChatID, messages[i].Timestamp, messages[i].MessageID)
		if err := storeInstance.Delete(key); err == nil {
			deleted++
		}
	}

	return deleted, nil
}

func CountMessages(chatID int64) int {
	storeInstance := store.GetInstance()
	iterator := storeInstance.Iterator()
	defer iterator.Release()

	prefix := []byte(fmt.Sprintf("message:%d:", chatID))
	count := 0

	for iterator.Next() {
		if bytes.HasPrefix(iterator.Key(), prefix) {
			count++
		}
	}

	return count
}

func GetMessageKey(chatID int64, timestamp int64, messageID int) []byte {
	return []byte(fmt.Sprintf("message:%d:%d:%d", chatID, timestamp, messageID))
}
