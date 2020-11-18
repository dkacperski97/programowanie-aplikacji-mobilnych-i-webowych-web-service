package auth

import (
	"context"
	"errors"
	"log"
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

func CreateUser(login, password, passwordConfirmation, email, firstname, lastname, address string) (*User, error) {
	isValid := IsValid(login, password, passwordConfirmation, email, firstname, lastname, address)
	if isValid == false {
		return nil, errors.New("User is not valid")
	}
	u := new(User)
	u.Login = login
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	u.PasswordHash = passwordHash
	u.Email = email
	u.Firstname = firstname
	u.Lastname = lastname
	u.Address = address
	return u, nil
}

func Verify(client *redis.Client, login, password string) bool {
	hash, err := client.HGet(context.Background(), "user:"+login, "passwordHash").Bytes()
	if err != nil {
		return false
	}
	err = bcrypt.CompareHashAndPassword(hash, []byte(password))
	return err == nil
}

func IsValid(login, password, passwordConfirmation, email, firstname, lastname, address string) bool {
	matched, err := regexp.MatchString(`[A-ZĄĆĘŁŃÓŚŹŻ][a-ząćęłńóśźż]+`, firstname)
	if err != nil {
		log.Panic("Incorrect regex pattern:", err)
	}
	if !matched {
		return false
	}

	matched, err = regexp.MatchString(`[A-ZĄĆĘŁŃÓŚŹŻ][a-ząćęłńóśźż]+`, lastname)
	if err != nil {
		log.Panic("Incorrect regex pattern:", err)
	}
	if !matched {
		return false
	}

	matched, err = regexp.MatchString(`[a-z]{3,12}`, login)
	if err != nil {
		log.Panic("Incorrect regex pattern:", err)
	}
	if !matched {
		return false
	}

	matched, err = regexp.MatchString(`[a-z]{3,12}`, password)
	if err != nil {
		log.Panic("Incorrect regex pattern:", err)
	}
	if !matched || password != passwordConfirmation {
		return false
	}

	matched, err = regexp.MatchString("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$", email)
	if err != nil {
		log.Panic("Incorrect regex pattern:", err)
	}
	if !matched {
		return false
	}

	if len(address) == 0 && len(address) > 200 {
		return false
	}

	return true
}

func (user *User) Save(client *redis.Client) {
	client.HSet(context.Background(), "user:"+user.Login, map[string]interface{}{
		"passwordHash": user.PasswordHash,
		"email":        user.Email,
		"firstname":    user.Firstname,
		"lastname":     user.Lastname,
		"address":      user.Address,
	})
}
