package services

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"time"
	"translate-server/datamodels"
)


type UserService interface {
	CheckUser(username, userPassword string) (*datamodels.User, bool)
	InsertUser(user datamodels.User) error
    QueryAdminUsers() ([]datamodels.User, error)
	QueryAllUsers() ([]datamodels.User, error)
	DeleteUserById(id int64) error
	UpdateUserPassword(user datamodels.User) error
    QueryUserByName(name string) (*datamodels.User, error)
}

func NewUserService() UserService  {
	return &userService{}
}

type userService struct {

}

func (u *userService) CheckUser(username, userPassword string) (*datamodels.User, bool){
	if username == "" || userPassword == "" {
		return nil, false
	}
	row := db.QueryRow("select * from tbl_user where Username = ?", username)
	var user datamodels.User
	err := row.Scan(&user.Id, &user.Username, &user.HashedPassword, &user.IsSuper, &user.CreatedAt)
	if err != nil {
		return nil, false
	}
	if ok, _ := datamodels.ValidatePassword(userPassword, user.HashedPassword); ok {
		return &user, true
	}
	return &user, false
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
	_, err = stmt.Exec(user.HashedPassword, user.Id)
	tx.Commit()
	if err != nil {
		log.Error(err)
		return err
	}
	return nil
}


func (u *userService) InsertUser(user datamodels.User) error {
	tx, _ := db.Begin()
	sql := fmt.Sprintf("INSERT INTO tbl_user(Username, HashedPassword, IsSuper) VALUES(?,?,?);")
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
	sql := fmt.Sprintf("SELECT Id, Username, IsSuper, CreatedAt FROM tbl_user where IsSuper=1")
	rows, err := db.Query(sql)
	defer rows.Close()
	if err != nil {
		log.Error(err)
		return nil, err
	}
	var users []datamodels.User
	for rows.Next() {
		user := datamodels.User{}
		var t time.Time
		err := rows.Scan(&user.Id, &user.Username, &user.IsSuper, &t)
		user.CreatedAt = t.Format("2006-01-02 15:04:05")
		if err != nil {
			log.Error(err)
		}
		users = append(users, user)
	}
	return users, nil
}

func (u *userService) QueryUserByName(name string) (*datamodels.User, error) {
	sql := fmt.Sprintf("SELECT Id, Username, IsSuper, CreatedAt FROM tbl_user where Username=?")
	row := db.QueryRow(sql, name)
	user := datamodels.User{}
	var t time.Time
	err := row.Scan(&user.Id, &user.Username, &user.IsSuper, &t)
	if err != nil {
		return nil, nil
	}
	user.CreatedAt = t.Format("2006-01-02 15:04:05")
	return &user, nil
}

func (u *userService) QueryAllUsers() ([]datamodels.User, error) {
	sql := fmt.Sprintf("SELECT Id, Username, IsSuper, CreatedAt FROM tbl_user where IsSuper=0")
	rows, err := db.Query(sql)
	defer rows.Close()
	if err != nil {
		log.Error(err)
		return nil, err
	}
	var users []datamodels.User
	for rows.Next() {
		user := datamodels.User{}
		var t time.Time
		err := rows.Scan(&user.Id, &user.Username, &user.IsSuper, &t)
		user.CreatedAt = t.Format("2006-01-02 15:04:05")
		if err != nil {
			log.Error(err)
			continue
		}
		users = append(users, user)
	}
	return users, nil
}
