package services

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	log "github.com/sirupsen/logrus"
)

var db *sql.DB

const initSql = `
CREATE TABLE IF NOT EXISTS "tbl_user" (
   "Id"	INTEGER PRIMARY KEY AUTOINCREMENT,
   "Username"	TEXT NOT NULL UNIQUE,
   "HashedPassword"	BLOB NOT NULL,
   "IsSuper" TINYINT,
   "CreatedAt" DATETIME DEFAULT (datetime('now','localtime'))
);

CREATE TABLE IF NOT EXISTS "tbl_record" (
  "Id"	INTEGER PRIMARY KEY AUTOINCREMENT,
  "Sha1"	TEXT,
  "Content"	TEXT,
  "ContentType" TEXT,
  "TransType" INTEGER,
  "OutputContent" TEXT,
  "SrcLang" TEXT,
  "DesLang" TEXT,
  "FileName" TEXT,
  "DirRandId" TEXT,
  "State" INTEGER,
  "StateDescribe" TEXT,
  "Error" TEXT,
  "UserId" INTEGER,
  "CreateAt" DATETIME DEFAULT (datetime('now','localtime'))
);
CREATE INDEX IF NOT EXISTS tbl_record_sha1_idx ON tbl_record(Sha1);

CREATE TABLE IF NOT EXISTS "tbl_component" (
    "ImageName" TEXT,
    "FileName" TEXT,
    "FilePath" text,
    "Version" TEXT,
    "FileMd5" TEXT,
    "CreateAt" DATETIME DEFAULT (datetime('now','localtime'))
);
`

func init() {
	var filename = "translate.db"
	db, _ = sql.Open("sqlite3", filename)
	_, err := db.Exec(initSql)
	if err != nil {
		log.Errorln(err)
		return
	}
}
