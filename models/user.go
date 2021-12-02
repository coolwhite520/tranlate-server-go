package models

import (
	"fmt"
	log "github.com/sirupsen/logrus"
)

const (
	TableName = "tbl_user"
)


func InsertUser(user User) error {
	tx, _ := db.Begin()
	sql := fmt.Sprintf("INSERT OR REPLACE INTO %s('Name', 'Password', 'IsSuper') VALUES(?,?,?);", TableName)
	stmt, err := tx.Prepare(sql)
	if err != nil {
		log.Error(err)
		return err
	}
	_, err = stmt.Exec(user.Name, user.Password, user.IsSuper)
	tx.Commit()
	if err != nil {
		log.Error(err)
		return err
	}
	return nil
}

func QueryAllUsers() ([]User, error) {
	sql := fmt.Sprintf("SELECT * FROM %s", TableName)
	rows, err := db.Query(sql)
	if err != nil {
		log.Error(err)
	}
	var users []User
	for rows.Next() {
		user := User{}
		err := rows.Scan(&user.Id, &user.Name, &user.Password, &user.IsSuper)
		if err != nil {
			log.Fatal(err)
		}
		users = append(users, user)
	}
	return users, nil
}
