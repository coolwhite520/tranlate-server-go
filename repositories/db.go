package repositories

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"translate-server/models"
)

var db *sql.DB

func init() {
	f, err := os.Open("./db.sqlite.sql")
	if err != nil {
		log.Errorln(err)
		return
	}
	defer f.Close()
	lines, _ := ioutil.ReadAll(f)
	var filename = "db.sqlite"
	db, _ = sql.Open("sqlite3", filename)
	_, err = db.Exec(string(lines[:]))
	if err != nil {
		log.Errorln(err)
		return
	}
	users, err := QueryAllUsers()
	if users == nil {
		password, _ := models.GeneratePassword("admin")
		InsertUser(models.User{
			Username:     "admin",
			HashedPassword: password,
			IsSuper:  true,
		})
	}
}
