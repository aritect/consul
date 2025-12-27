package model

import (
	"consul-telegram-bot/internal/store"
	"fmt"

	"github.com/vmihailenco/msgpack/v5"
)

type LLMContext struct {
	ChatID  int64  `msgpack:"chat_id"`
	Context string `msgpack:"context"`
}

func NewLLMContext(chatID int64, context string) (*LLMContext, error) {
	ctx := &LLMContext{
		ChatID:  chatID,
		Context: context,
	}

	if err := ctx.Save(); err != nil {
		return nil, err
	}

	return ctx, nil
}

func (ctx *LLMContext) Save() error {
	key := GetLLMContextKey(ctx.ChatID)
	data, err := msgpack.Marshal(ctx)
	if err != nil {
		return err
	}

	storeInstance := store.GetInstance()
	return storeInstance.Put(key, data)
}

func FindLLMContext(chatID int64) (*LLMContext, error) {
	storeInstance := store.GetInstance()
	key := GetLLMContextKey(chatID)
	data, err := storeInstance.Get(key)
	if err != nil {
		return nil, err
	}

	var ctx LLMContext
	if err := msgpack.Unmarshal(data, &ctx); err != nil {
		return nil, err
	}

	return &ctx, nil
}

func DeleteLLMContext(chatID int64) error {
	storeInstance := store.GetInstance()
	key := GetLLMContextKey(chatID)
	return storeInstance.Delete(key)
}

func GetLLMContextKey(chatID int64) []byte {
	return []byte(fmt.Sprintf("llm_context:%d", chatID))
}
