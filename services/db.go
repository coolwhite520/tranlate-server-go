package services

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	log "github.com/sirupsen/logrus"
	"time"
	"translate-server/config"
	"translate-server/datamodels"
)

var db *sql.DB

var SqlArr = []string{
	`CREATE DATABASE IF NOT EXISTS translate_db DEFAULT CHARACTER SET utf8 COLLATE utf8_general_ci;`,
	`use translate_db;`,
	`CREATE TABLE IF NOT EXISTS tbl_user (
   Id int(11) NOT NULL AUTO_INCREMENT,
   Username	VARCHAR(255) UNIQUE,
   HashedPassword	BLOB NOT NULL,
   IsSuper TINYINT,
   CreatedAt DATETIME NULL DEFAULT CURRENT_TIMESTAMP,
   PRIMARY KEY (Id)
)ENGINE=InnoDB DEFAULT CHARSET=utf8;`,
	`CREATE TABLE IF NOT EXISTS tbl_record (
  Id int(11) NOT NULL AUTO_INCREMENT,
  Sha1	VARCHAR(255),
  Content	TEXT,
  ContentType TEXT,
  TransType INTEGER,
  OutputContent TEXT,
  SrcLang TEXT,
  DesLang TEXT,
  FileName TEXT,
  DirRandId TEXT,
  State INTEGER,
  StateDescribe TEXT,
  Error TEXT,
  UserId INTEGER,
  CreateAt DATETIME NULL DEFAULT CURRENT_TIMESTAMP,
  INDEX(Sha1),
  PRIMARY KEY (Id)
)ENGINE=InnoDB DEFAULT CHARSET=utf8;`,
}

func InitDb() {
	var hostPort string
	list, err := config.GetInstance().GetComponentList(false)
	if err != nil {
		panic(err)
	}
	for _, v := range list {
		if v.ImageName == "mysql" {
			hostPort = v.HostPort
			break
		}
	}
	time.Sleep(10 * time.Second)
	dataSourceName := fmt.Sprintf("root:%s@tcp(127.0.0.1:%s)/?charset=utf8&parseTime=True", datamodels.MysqlPassword, hostPort)
	db, err = sql.Open("mysql", dataSourceName)
	if err != nil {
		log.Error(err)
		panic(err)
	}
	for _, v := range SqlArr {
		_, err = db.Exec(v)
		if err != nil {
			log.Error(err)
			break
		}

	}
	db.Close()
	// 重新建立一个链接，链接到translate_db数据库，就不需要切换操作了
	dataSourceName = fmt.Sprintf("root:%s@tcp(127.0.0.1:%s)/translate_db?charset=utf8&parseTime=True", datamodels.MysqlPassword, hostPort)
	db, err = sql.Open("mysql", dataSourceName)
	if err != nil {
		log.Errorln(err)
		panic(err)
	}
	service := NewUserService()
	users, _ := service.QueryAdminUsers()
	if users == nil {
		password, _ := datamodels.GeneratePassword("admin")
		service.InsertUser(datamodels.User{
			Username:       fmt.Sprintf("admin"),
			HashedPassword: password,
			IsSuper:        true,
		})
	}

}
