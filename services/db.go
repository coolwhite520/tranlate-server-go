package services

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	log "github.com/sirupsen/logrus"
)

var db *sql.DB

var SqlArr = []string {
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


func init() {
	var err error
	db, err = sql.Open("mysql", "root:112233@tcp(127.0.0.1:3306)/?charset=utf8&parseTime=True")
	if err != nil {
		log.Errorln(err)
		return
	}
	for _,v := range SqlArr {
		exec, err := db.Exec(v)
		if err != nil {
			panic(err)
		}
		log.Println(exec)
	}
}
