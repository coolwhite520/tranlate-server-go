package datamodels

import (
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Id             int64  `json:"id" form:"id"`
	Username       string `json:"username"`
	HashedPassword []byte `json:"-" form:"-"`
	IsSuper        bool   `json:"is_super"`
	CreatedAt      string `json:"created_at"`
}

func (u User) IsValid() bool {
	return u.Id > 0
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
