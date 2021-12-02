package repositories

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"translate-server/models"
)

const (
	TableName = "tbl_user"
)


func InsertUser(user models.User) error {
	tx, _ := db.Begin()
	sql := fmt.Sprintf("INSERT OR REPLACE INTO %s('Username', 'HashedPassword', 'IsSuper') VALUES(?,?,?);", TableName)
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

func QueryAllUsers() ([]models.User, error) {
	sql := fmt.Sprintf("SELECT ID, Username,IsSuper FROM %s", TableName)
	rows, err := db.Query(sql)
	if err != nil {
		log.Error(err)
	}
	var users []models.User
	for rows.Next() {
		user := models.User{}
		err := rows.Scan(&user.ID, &user.Username, &user.IsSuper)
		if err != nil {
			log.Fatal(err)
		}
		users = append(users, user)
	}
	return users, nil
}
