package models

import (
	"context"
	"errors"
	"regexp"
	"strconv"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
)

type Label struct {
	ID        string
	Sender    string
	Recipient string
	Locker    string
	Size      int
}

func CreateLabel(sender, recipient, locker, size string) (*Label, error, error) {
	sizeValue, err := strconv.Atoi(size)
	if err != nil {
		return nil, errors.New("Niepoprawny rozmiar paczki"), nil
	}
	validationErr, err := IsLabelValid(sender, recipient, locker, sizeValue)
	if validationErr != nil {
		return nil, validationErr, nil
	}
	if err != nil {
		return nil, nil, err
	}
	l := new(Label)
	l.Sender = sender
	l.Recipient = recipient
	l.Locker = locker
	l.Size = sizeValue
	return l, nil, nil
}

func IsLabelValid(sender, recipient, locker string, size int) (error, error) {
	matched, err := regexp.MatchString(`[a-z]{3,12}`, sender)
	if err != nil {
		return nil, err
	}
	if !matched {
		return errors.New("Niepoprawny nadawca"), nil
	}

	if len(recipient) == 0 && len(recipient) > 100 {
		return errors.New("Niepoprawny adresat"), nil
	}

	matched, err = regexp.MatchString(`[A-Z0-9]{9}`, locker)
	if err != nil {
		return nil, err
	}
	if !matched {
		return errors.New("Niepoprawny identyfikator skrytki"), nil
	}

	if size < 1 && size > 8000 {
		return errors.New("Niepoprawny rozmiar paczki"), nil
	}

	return nil, nil
}

func (label *Label) Save(client *redis.Client) error {
	id := uuid.New().String()
	err := client.HSet(context.Background(), "label:"+id, map[string]interface{}{
		"sender":    label.Sender,
		"recipient": label.Recipient,
		"locker":    label.Locker,
		"size":      label.Size,
	}).Err()
	if err != nil {
		return err
	}
	err = client.SAdd(context.Background(), "user:"+label.Sender+":labels", id).Err()
	return err
}

func GetLabelsBySender(client *redis.Client, sender string) ([]Label, error) {
	val, err := client.SMembers(context.Background(), "user:"+sender+":labels").Result()
	if err != nil {
		return nil, err
	}
	labels := []Label{}
	for _, id := range val {
		val, err := client.HGetAll(context.Background(), "label:"+id).Result()
		if err != nil {
			return nil, err
		}
		size, err := strconv.Atoi(val["size"])
		if err != nil {
			return nil, err
		}
		label := Label{id, val["sender"], val["recipient"], val["locker"], size}
		labels = append(labels, label)
	}
	return labels, nil
}

func RemoveLabel(client *redis.Client, sender, labelID string) error {
	err := client.Del(context.Background(), "label:"+labelID).Err()
	if err != nil {
		return err
	}
	err = client.SRem(context.Background(), "user:"+sender+":labels", labelID).Err()
	return err
}
