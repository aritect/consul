package model

import (
	"bytes"
	"consul-telegram-bot/internal/store"
	"fmt"
	"sync"

	"github.com/vmihailenco/msgpack/v5"
)

var mu sync.RWMutex

type RecipientType string

const (
	RecipientPrivate    RecipientType = "private"
	RecipientGroup      RecipientType = "group"
	RecipientSuperGroup RecipientType = "supergroup"
	RecipientChannel    RecipientType = "channel"
)

type Recipient struct {
	Id                  int64
	Type                RecipientType
	ThreadId            int
	AritectBuysThreadId int
	Receiving           int64
}

func NewRecipient(id int64, recipientType RecipientType, threadId int) (*Recipient, error) {
	mu.Lock()
	defer mu.Unlock()

	if r, err := FindRecipient(id); err == nil {
		r.Type = recipientType

		r.Write()

		return r, nil
	}

	r := &Recipient{
		Id:        id,
		ThreadId:  threadId,
		Type:      recipientType,
		Receiving: 1,
	}

	key := GetRecipientKey(id)
	bytes, err := msgpack.Marshal(r)
	if err != nil {
		return r, err
	}

	storeInstance := store.GetInstance()
	err = storeInstance.Put(key, bytes)
	if err != nil {
		return r, err
	}

	return r, nil
}

func (r Recipient) Recipient() string {
	return fmt.Sprintf("%d", r.Id)
}

func (r *Recipient) Write() error {
	key := GetRecipientKey(r.Id)
	bytes, err := msgpack.Marshal(r)
	if err != nil {
		return err
	}

	storeInstance := store.GetInstance()
	err = storeInstance.Put(key, bytes)
	if err != nil {
		return err
	}

	return nil
}

func (r *Recipient) EnableReceiving() {
	r.Receiving = 1
}

func (r *Recipient) DisableReceiving() {
	r.Receiving = -1
}

func (r *Recipient) IsEnabledReceiving() bool {
	return r.Receiving != -1
}

func (r *Recipient) DefineThreadId(threadId int) {
	r.ThreadId = threadId
}

func (r *Recipient) DefineThreadIdForSignalType(signalType SignalType, threadId int) {
	switch signalType {
	case SignalTypeAritectBuys:
		r.AritectBuysThreadId = threadId
	}
}

func (r *Recipient) GetThreadIdForSignalType(signalType SignalType) int {
	switch signalType {
	case SignalTypeAritectBuys:
		return r.AritectBuysThreadId
	default:
		return 0
	}
}

func (r *Recipient) DeleteSelf() error {
	storeInstance := store.GetInstance()
	err := storeInstance.Delete(GetRecipientKey(r.Id))
	if err != nil {
		return err
	}

	return nil
}

func FindRecipient(id int64) (*Recipient, error) {
	storeInstance := store.GetInstance()
	key := GetRecipientKey(id)
	bsValue, err := storeInstance.Get(key)
	if err != nil {
		return nil, err
	}

	return UnmarshalRecipient(bsValue)
}

func FindAllRecipientsByIds(ids []int64) []*Recipient {
	recipients := make([]*Recipient, 0)
	for _, id := range ids {
		r, err := FindRecipient(id)
		if err != nil {
			continue
		}

		recipients = append(recipients, r)
	}
	return recipients
}

func FindAllRecipients() []*Recipient {
	recipients := make([]*Recipient, 0)
	IterateRecipients(func(r *Recipient) {
		recipients = append(recipients, r)
	})
	return recipients
}

func UnmarshalRecipient(b []byte) (*Recipient, error) {
	var r *Recipient
	if err := msgpack.Unmarshal(b, &r); err != nil {
		return nil, err
	}

	return r, nil
}

func IterateRecipients(fn func(*Recipient)) {
	storeInstance := store.GetInstance()
	iterator := storeInstance.Iterator()
	defer iterator.Release()

	prefix := []byte("recipient:")
	for iterator.Next() {
		key := iterator.Key()
		if !bytes.HasPrefix(key, prefix) {
			continue
		}
		if r, err := UnmarshalRecipient(iterator.Value()); err == nil {
			fn(r)
		}
	}
}

func GetRecipientKey(id int64) []byte {
	return []byte(fmt.Sprintf("recipient:%d", id))
}
