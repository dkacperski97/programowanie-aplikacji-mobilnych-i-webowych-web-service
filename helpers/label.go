package helpers

import (
	"context"
	"strconv"
	"strings"

	models "github.com/dkacperski97/programowanie-aplikacji-mobilnych-i-webowych-models"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
)

func SaveLabel(client *redis.Client, label *models.Label) error {
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
	if err == nil {
		label.ID = models.LabelID(id)
	}
	return err
}

func GetLabelsBySender(client *redis.Client, sender string) ([]models.Label, error) {
	val, err := client.SMembers(context.Background(), "user:"+sender+":labels").Result()
	if err != nil {
		return nil, err
	}
	labels := []models.Label{}
	for _, id := range val {
		val, err := client.HGetAll(context.Background(), "label:"+id).Result()
		if err != nil {
			return nil, err
		}
		size, err := strconv.Atoi(val["size"])
		if err != nil {
			return nil, err
		}
		label := models.Label{
			ID:             models.LabelID(id),
			Sender:         val["sender"],
			Recipient:      val["recipient"],
			Locker:         val["locker"],
			Size:           size,
			AssignedParcel: val["assignedParcel"],
		}
		labels = append(labels, label)
	}
	return labels, nil
}

func GetLabel(client *redis.Client, labelID string) (*models.Label, error) {
	val, err := client.HGetAll(context.Background(), "label:"+labelID).Result()
	if err != nil {
		return nil, err
	}
	size, err := strconv.Atoi(val["size"])
	if err != nil {
		return nil, err
	}
	label := models.Label{
		ID:             models.LabelID(labelID),
		Sender:         val["sender"],
		Recipient:      val["recipient"],
		Locker:         val["locker"],
		Size:           size,
		AssignedParcel: val["assignedParcel"],
	}
	return &label, nil
}

func GetLabels(client *redis.Client) ([]models.Label, error) {
	val, err := client.Keys(context.Background(), "label*").Result()
	if err != nil {
		return nil, err
	}
	labels := []models.Label{}
	for _, id := range val {
		val, err := client.HGetAll(context.Background(), id).Result()
		if err != nil {
			return nil, err
		}
		size, err := strconv.Atoi(val["size"])
		if err != nil {
			return nil, err
		}
		label := models.Label{
			ID:             models.LabelID(strings.Replace(id, "label:", "", 1)),
			Sender:         val["sender"],
			Recipient:      val["recipient"],
			Locker:         val["locker"],
			Size:           size,
			AssignedParcel: val["assignedParcel"],
		}
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
