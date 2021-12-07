package services

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"time"
	"translate-server/datamodels"
)


type UserService interface {
	CheckUser(username, userPassword string) (datamodels.User, bool)
	InsertUser(user datamodels.User) error
    QueryAdminUsers() ([]datamodels.User, error)
	QueryAllUsers() ([]datamodels.User, error)
	DeleteUserById(id int64) error
	UpdateUserPassword(user datamodels.User) error
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
func (u *userService) DeleteUserById(Id int64) error {
	tx, _ := db.Begin()
	sql := fmt.Sprintf("Delete From tbl_user where Id=?;")
	stmt, err := tx.Prepare(sql)
	if err != nil {
		log.Error(err)
		return err
	}
	_, err = stmt.Exec(Id)
	tx.Commit()
	if err != nil {
		log.Error(err)
		return err
	}
	return nil
}

func (u *userService) UpdateUserPassword(user datamodels.User) error {
	tx, _ := db.Begin()
	sql := fmt.Sprintf("UPDATE tbl_user set HashedPassword=? where Id=?;")
	stmt, err := tx.Prepare(sql)
	if err != nil {
		log.Error(err)
		return err
	}
	_, err = stmt.Exec(user.HashedPassword, user.ID)
	tx.Commit()
	if err != nil {
		log.Error(err)
		return err
	}
	return nil
}


func (u *userService) InsertUser(user datamodels.User) error {
	tx, _ := db.Begin()
	sql := fmt.Sprintf("INSERT INTO tbl_user('Username', 'HashedPassword', 'IsSuper') VALUES(?,?,?);")
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
func (u *userService) QueryAdminUsers() ([]datamodels.User, error) {
	sql := fmt.Sprintf("SELECT ID, Username, IsSuper, CreatedAt FROM tbl_user where IsSuper=1")
	rows, err := db.Query(sql)
	if err != nil {
		log.Error(err)
	}
	var users []datamodels.User
	for rows.Next() {
		user := datamodels.User{}
		var t time.Time
		err := rows.Scan(&user.ID, &user.Username, &user.IsSuper, &t)
		user.CreatedAt = t.Local().Format("2006-01-02 15:04:05")
		if err != nil {
			log.Fatal(err)
		}
		users = append(users, user)
	}
	return users, nil
}

func (u *userService) QueryAllUsers() ([]datamodels.User, error) {
	sql := fmt.Sprintf("SELECT ID, Username, IsSuper, CreatedAt FROM tbl_user where IsSuper=0")
	rows, err := db.Query(sql)
	if err != nil {
		log.Error(err)
	}
	var users []datamodels.User
	for rows.Next() {
		user := datamodels.User{}
		var t time.Time
		err := rows.Scan(&user.ID, &user.Username, &user.IsSuper, &t)
		user.CreatedAt = t.Local().Format("2006-01-02 15:04:05")
		if err != nil {
			log.Fatal(err)
		}
		users = append(users, user)
	}
	return users, nil
}
