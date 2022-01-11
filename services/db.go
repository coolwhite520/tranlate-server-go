package services

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	log "github.com/sirupsen/logrus"
	"reflect"
	"time"
	"translate-server/config"
	"translate-server/datamodels"
)

var db *sql.DB

var SqlArr = []string{
	`CREATE DATABASE IF NOT EXISTS translate_db DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;`,
	`use translate_db;`,
	`CREATE TABLE IF NOT EXISTS tbl_user (
	   Id int(11) NOT NULL AUTO_INCREMENT,
	   Username	VARCHAR(255) UNIQUE,
	   HashedPassword	BLOB NOT NULL,
	   IsSuper TINYINT,
	   Mark TEXT,
	   CreatedAt DATETIME NULL DEFAULT CURRENT_TIMESTAMP,
	   PRIMARY KEY (Id)
	)ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;`,
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
	)ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;`,
	`CREATE TABLE IF NOT EXISTS tbl_user_operator (
	   Id int(11) NOT NULL AUTO_INCREMENT,
	   UserId int(11),
       Ip TEXT,
       Operator TEXT,
       CreateAt DATETIME NULL DEFAULT CURRENT_TIMESTAMP,
	   PRIMARY KEY (Id)
	)ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;`,
	`CREATE TABLE IF NOT EXISTS tbl_ips (
	   Id int(11) NOT NULL AUTO_INCREMENT,
       Ip TEXT,
       Type VARCHAR(255),
       CreateAt DATETIME NULL DEFAULT CURRENT_TIMESTAMP,
	   PRIMARY KEY (Id)
	)ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;`,
}

func InitDb() {
	service := NewUserService()
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
	dataSourceName := fmt.Sprintf("root:%s@tcp(127.0.0.1:%s)/?charset=utf8&parseTime=True", datamodels.MysqlPassword, hostPort)
	for i:=0; i < 100; i++ {
		time.Sleep(1 * time.Second)
		db, err = sql.Open("mysql", dataSourceName)
		if err != nil {
			log.Error(fmt.Sprintf("attempt to connect mysql port:%s, ", hostPort), err.Error())
			continue
		}
		err = db.Ping()
		if err != nil{
			log.Error(fmt.Sprintf("attempt to connect mysql port:%s, ", hostPort), err.Error())
			continue
		}
		break
	}

	count, err := service.QueryTableFieldCount("translate_db", "tbl_user")
	if err != nil {
		log.Error(err)
	}
	var user datamodels.User
	typeOfUser := reflect.TypeOf(user)
	userFieldCount := typeOfUser.NumField()
	if count != userFieldCount {
		log.Info("OldUserTblFieldCount:", count," NewUserTblFieldCount:", userFieldCount)
		err := service.DropDatabase("translate_db")
		if err != nil {
			log.Error(err)
		}
	}
	count, err = service.QueryTableFieldCount("translate_db", "tbl_record")
	if err != nil {
		log.Error(err)
	}
	var record datamodels.Record
	typeOfRecord := reflect.TypeOf(record)
	recordFieldCount := typeOfRecord.NumField()
	if count != recordFieldCount {
		log.Info("OldRecordTblFieldCount:", count," NewRecordTblFieldCount:", recordFieldCount)
		err := service.DropDatabase("translate_db")
		if err != nil {
			log.Error(err)
		}
	}
	// 数据库和表的初始化
	for _, v := range SqlArr {
		_, err = db.Exec(v)
		if err != nil {
			log.Error(err)
			panic(err)
		}

	}
	// 重新建立一个链接，链接到translate_db数据库，就不需要切换操作了
	dataSourceName = fmt.Sprintf("root:%s@tcp(127.0.0.1:%s)/translate_db?charset=utf8&parseTime=True", datamodels.MysqlPassword, hostPort)
	db, err = sql.Open("mysql", dataSourceName)
	if err != nil {
		log.Error(err)
		panic(err)
	}

	users, _ := service.QueryAdminUsers()
	if users == nil {
		password, _ := datamodels.GeneratePassword("admin")
		service.InsertUser(datamodels.User{
			Username:       fmt.Sprintf("admin"),
			HashedPassword: password,
			IsSuper:        true,
			Mark: "超级管理员",
		})
	}

}
