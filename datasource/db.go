package datasource

import (
	"database/sql"
	"io/ioutil"
	"os"
)

var filename = "db.sqlite"
var db, _ = sql.Open("sqlite3", filename)

func init()  {
	f, _ := os.Open("./db.sqlite.sql")
	defer f.Close()
	lines, _ := ioutil.ReadAll(f)
	db.Exec(string(lines[:]))
}