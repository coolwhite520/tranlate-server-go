package services

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"translate-server/datamodels"
)


type UserService interface {
	CheckUser(username, userPassword string) (datamodels.User, bool)
	InsertUser(user datamodels.User) error
	QueryAllUsers() ([]datamodels.User, error)
}

func NewUserService() UserService  {
	return &userService{}
}

type userService struct {

}

func (u *userService) CheckUser(username, userPassword string) (datamodels.User, bool){
	if username == "" || userPassword == "" {
		return datamodels.User{}, false
	}
	row := db.QueryRow("select * from tbl_user where Username = ?", username)
	var user datamodels.User
	err := row.Scan(&user.ID, &user.Username, &user.HashedPassword, &user.IsSuper, &user.CreatedAt)
	if err != nil {
		return datamodels.User{}, false
	}
	if ok, _ := datamodels.ValidatePassword(userPassword, user.HashedPassword); ok {
		return user, true
	}
	return datamodels.User{}, false

}

func (u *userService)InsertUser(user datamodels.User) error {
	tx, _ := db.Begin()
	sql := fmt.Sprintf("INSERT OR REPLACE INTO tbl_user('Username', 'HashedPassword', 'IsSuper') VALUES(?,?,?);")
	stmt, err := tx.Prepare(sql)
	if err != nil {
		log.Error(err)
		return err
	}
	_, err = stmt.Exec(user.Username, user.HashedPassword, user.IsSuper)
	tx.Commit()
	if err != nil {
		log.Error(err)
		return err
	}
	return nil
}

func (u *userService)QueryAllUsers() ([]datamodels.User, error) {
	sql := fmt.Sprintf("SELECT ID, Username,IsSuper FROM tbl_user")
	rows, err := db.Query(sql)
	if err != nil {
		log.Error(err)
	}
	var users []datamodels.User
	for rows.Next() {
		user := datamodels.User{}
		err := rows.Scan(&user.ID, &user.Username, &user.IsSuper)
		if err != nil {
			log.Fatal(err)
		}
		users = append(users, user)
	}
	return users, nil
}
