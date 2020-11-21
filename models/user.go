package models

import (
	"context"
	"errors"
	"regexp"

	"github.com/go-redis/redis/v8"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Login        string
	PasswordHash []byte
	Email        string
	Firstname    string
	Lastname     string
	Address      string
}

func CreateUser(login, password, email, firstname, lastname, address string) (*User, error, error) {
	validationErr, err := IsUserValid(login, password, email, firstname, lastname, address)
	if validationErr != nil {
		return nil, validationErr, nil
	}
	if err != nil {
		return nil, nil, err
	}
	u := new(User)
	u.Login = login
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, nil, err
	}
	u.PasswordHash = passwordHash
	u.Email = email
	u.Firstname = firstname
	u.Lastname = lastname
	u.Address = address
	return u, nil, nil
}

func Verify(client *redis.Client, login, password string) (bool, error) {
	res1, err := client.HExists(context.Background(), "user:"+login, "passwordHash").Result()
	if err != nil {
		return false, err
	}
	if res1 == false {
		return false, nil
	}
	res2, _ := client.HGet(context.Background(), "user:"+login, "passwordHash").Result()
	err = bcrypt.CompareHashAndPassword([]byte(res2), []byte(password))
	return err == nil, nil
}

func IsUserValid(login, password, email, firstname, lastname, address string) (error, error) {
	matched, err := regexp.MatchString(`[A-ZĄĆĘŁŃÓŚŹŻ][a-ząćęłńóśźż]+`, firstname)
	if err != nil {
		return nil, err
	}
	if !matched {
		return errors.New("Niepoprawne imię"), nil
	}

	matched, err = regexp.MatchString(`[A-ZĄĆĘŁŃÓŚŹŻ][a-ząćęłńóśźż]+`, lastname)
	if err != nil {
		return nil, err
	}
	if !matched {
		return errors.New("Niepoprawne nazwisko"), nil
	}

	matched, err = regexp.MatchString(`[a-z]{3,12}`, login)
	if err != nil {
		return nil, err
	}
	if !matched {
		return errors.New("Niepoprawny login"), nil
	}

	matched, err = regexp.MatchString(`.{8,}`, password)
	if err != nil {
		return nil, err
	}
	if !matched {
		return errors.New("Niepoprawne hasło"), nil
	}

	matched, err = regexp.MatchString("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$", email)
	if err != nil {
		return nil, err
	}
	if !matched {
		return errors.New("Niepoprawny email"), nil
	}

	if len(address) == 0 && len(address) > 200 {
		return errors.New("Niepoprawny adres"), nil
	}

	return nil, nil
}

func (user *User) Save(client *redis.Client) error {
	err := client.HSet(context.Background(), "user:"+user.Login, map[string]interface{}{
		"passwordHash": user.PasswordHash,
		"email":        user.Email,
		"firstname":    user.Firstname,
		"lastname":     user.Lastname,
		"address":      user.Address,
	}).Err()
	return err
}
