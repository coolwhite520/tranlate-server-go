package models

import (
	"golang.org/x/crypto/bcrypt"
	"time"
)

type User struct {
	ID             int64     `json:"id" form:"id"`
	Username       string    `json:"username" form:"username"`
	HashedPassword []byte    `json:"-" form:"-"`
	IsSuper		   bool     `json:"is_super"`
	CreatedAt      time.Time `json:"created_at" form:"created_at"`
}

func (u User) IsValid() bool {
	return u.ID > 0
}

func GeneratePassword(userPassword string) ([]byte, error) {
	return bcrypt.GenerateFromPassword([]byte(userPassword), bcrypt.DefaultCost)
}

func ValidatePassword(userPassword string, hashed []byte) (bool, error) {
	if err := bcrypt.CompareHashAndPassword(hashed, []byte(userPassword)); err != nil {
		return false, err
	}
	return true, nil
}
