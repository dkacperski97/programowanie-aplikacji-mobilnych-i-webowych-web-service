package models

import (
	"context"
	"errors"
	"strings"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
)

const (
	ParcelStatusOnTheWay  string = "on_the_way"
	ParcelStatusDelivered string = "delivered"
	ParcelStatusReceived  string = "received"
)

type Parcel struct {
	ID      string `json:"id"`
	LabelID string `json:"labelId"`
	Status  string `json:"status"`
}

func CreateParcel(labelID, status string) (*Parcel, error, error) {
	validationErr, err := IsParcelValid(labelID, status)
	if validationErr != nil {
		return nil, validationErr, nil
	}
	if err != nil {
		return nil, nil, err
	}
	p := new(Parcel)
	p.LabelID = labelID
	p.Status = status
	return p, nil, nil
}

func IsParcelValid(labelID, status string) (error, error) {
	if status != ParcelStatusOnTheWay && status != ParcelStatusDelivered && status != ParcelStatusReceived {
		return nil, errors.New("Status is not valid")
	}
	_, err := uuid.Parse(labelID)
	if err != nil {
		return err, nil
	}

	return nil, nil
}

func (parcel *Parcel) Save(client *redis.Client) error {
	val, err := client.HGet(context.Background(), "label:"+parcel.LabelID, "assignedParcel").Result()
	if err != nil && err != redis.Nil {
		return err
	}
	if val != "" {
		return errors.New("Label already assigned")
	}
	id := uuid.New().String()
	err = client.HSet(context.Background(), "parcel:"+id, map[string]interface{}{
		"labelId": parcel.LabelID,
		"status":  parcel.Status,
	}).Err()
	if err != nil {
		return err
	}

	err = client.HSet(context.Background(), "label:"+parcel.LabelID, "assignedParcel", id).Err()
	return err
}

func GetParcels(client *redis.Client) ([]Parcel, error) {
	val, err := client.Keys(context.Background(), "parcel*").Result()
	if err != nil {
		return nil, err
	}
	parcels := []Parcel{}
	for _, id := range val {
		val, err := client.HGetAll(context.Background(), id).Result()
		if err != nil {
			return nil, err
		}
		parcel := Parcel{
			ID:      strings.Replace(id, "parcel:", "", 1),
			LabelID: val["labelId"],
			Status:  val["status"],
		}
		parcels = append(parcels, parcel)
	}
	return parcels, nil
}

func (parcel *Parcel) UpdateStatus(client *redis.Client) error {
	err := client.HSet(context.Background(), "parcel:"+parcel.ID, "status", parcel.Status).Err()
	return err
}
