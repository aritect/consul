package model

import (
	"bytes"
	"consul-telegram-bot/internal/store"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/vmihailenco/msgpack/v5"
)

var ratingMu sync.RWMutex

type UserRating struct {
	ChatID         int64  `msgpack:"chat_id"`
	UserID         int64  `msgpack:"user_id"`
	Username       string `msgpack:"username"`
	DisplayName    string `msgpack:"display_name"`
	Points         int    `msgpack:"points"`
	LastUpReceived int64  `msgpack:"last_up_received"`
	UpdatedAt      int64  `msgpack:"updated_at"`
}

type UpVote struct {
	ChatID    int64 `msgpack:"chat_id"`
	VoterID   int64 `msgpack:"voter_id"`
	TargetID  int64 `msgpack:"target_id"`
	Timestamp int64 `msgpack:"timestamp"`
}

func GetOrCreateUserRating(chatID, userID int64, username, displayName string) (*UserRating, error) {
	ratingMu.Lock()
	defer ratingMu.Unlock()

	rating, err := findUserRating(chatID, userID)
	if err == nil {
		if rating.Username != username || rating.DisplayName != displayName {
			rating.Username = username
			rating.DisplayName = displayName
			rating.save()
		}
		return rating, nil
	}

	rating = &UserRating{
		ChatID:      chatID,
		UserID:      userID,
		Username:    username,
		DisplayName: displayName,
		Points:      0,
		UpdatedAt:   time.Now().Unix(),
	}

	if err := rating.save(); err != nil {
		return nil, err
	}

	return rating, nil
}

func (r *UserRating) AddPoint() error {
	ratingMu.Lock()
	defer ratingMu.Unlock()

	r.Points++
	r.LastUpReceived = time.Now().Unix()
	r.UpdatedAt = time.Now().Unix()
	return r.save()
}

func (r *UserRating) save() error {
	key := getUserRatingKey(r.ChatID, r.UserID)
	data, err := msgpack.Marshal(r)
	if err != nil {
		return err
	}

	storeInstance := store.GetInstance()
	return storeInstance.Put(key, data)
}

func findUserRating(chatID, userID int64) (*UserRating, error) {
	storeInstance := store.GetInstance()
	key := getUserRatingKey(chatID, userID)
	data, err := storeInstance.Get(key)
	if err != nil {
		return nil, err
	}

	var rating UserRating
	if err := msgpack.Unmarshal(data, &rating); err != nil {
		return nil, err
	}

	return &rating, nil
}

func GetTopRatings(chatID int64, limit int) ([]*UserRating, error) {
	ratingMu.RLock()
	defer ratingMu.RUnlock()

	storeInstance := store.GetInstance()
	iterator := storeInstance.Iterator()
	defer iterator.Release()

	prefix := []byte(fmt.Sprintf("rating:%d:", chatID))
	ratings := make([]*UserRating, 0)

	for iterator.Next() {
		key := iterator.Key()
		if !bytes.HasPrefix(key, prefix) {
			continue
		}

		var rating UserRating
		if err := msgpack.Unmarshal(iterator.Value(), &rating); err != nil {
			continue
		}

		if rating.Points > 0 {
			ratings = append(ratings, &rating)
		}
	}

	sort.Slice(ratings, func(i, j int) bool {
		return ratings[i].Points > ratings[j].Points
	})

	if len(ratings) > limit {
		ratings = ratings[:limit]
	}

	return ratings, nil
}

func CanUserVote(chatID, voterID, targetID int64) (bool, error) {
	ratingMu.RLock()
	defer ratingMu.RUnlock()

	storeInstance := store.GetInstance()
	key := getUpVoteKey(chatID, voterID, targetID)
	data, err := storeInstance.Get(key)
	if err != nil {
		return true, nil
	}

	var vote UpVote
	if err := msgpack.Unmarshal(data, &vote); err != nil {
		return true, nil
	}

	hourAgo := time.Now().Add(-1 * time.Hour).Unix()
	return vote.Timestamp < hourAgo, nil
}

func RecordVote(chatID, voterID, targetID int64) error {
	ratingMu.Lock()
	defer ratingMu.Unlock()

	vote := &UpVote{
		ChatID:    chatID,
		VoterID:   voterID,
		TargetID:  targetID,
		Timestamp: time.Now().Unix(),
	}

	key := getUpVoteKey(chatID, voterID, targetID)
	data, err := msgpack.Marshal(vote)
	if err != nil {
		return err
	}

	storeInstance := store.GetInstance()
	return storeInstance.Put(key, data)
}

func getUserRatingKey(chatID, userID int64) []byte {
	return []byte(fmt.Sprintf("rating:%d:%d", chatID, userID))
}

func getUpVoteKey(chatID, voterID, targetID int64) []byte {
	return []byte(fmt.Sprintf("upvote:%d:%d:%d", chatID, voterID, targetID))
}
