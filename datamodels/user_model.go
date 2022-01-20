package datamodels

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"time"
	"translate-server/structs"
)

func CheckUser(username, userPassword string) (*structs.User, bool) {
	if username == "" || userPassword == "" {
		return nil, false
	}
	row := db.QueryRow("select * from tbl_user where Username = ?", username)
	var user structs.User
	err := row.Scan(&user.Id, &user.Username, &user.HashedPassword, &user.IsSuper, &user.Mark, &user.CreatedAt)
	if err != nil {
		return nil, false
	}
	if ok, _ := structs.ValidatePassword(userPassword, user.HashedPassword); ok {
		return &user, true
	}
	return &user, false
}

func DeleteUserById(Id int64) error {
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

func UpdateUserPassword(user structs.User) error {
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

func UpdateUserMark(user structs.User) error {
	tx, _ := db.Begin()
	sql := fmt.Sprintf("UPDATE tbl_user set Mark=? where Id=?;")
	stmt, err := tx.Prepare(sql)
	if err != nil {
		log.Error(err)
		return err
	}
	_, err = stmt.Exec(user.Mark, user.Id)
	tx.Commit()
	if err != nil {
		log.Error(err)
		return err
	}
	return nil
}

func InsertUser(user structs.User) error {
	tx, _ := db.Begin()
	sql := fmt.Sprintf("INSERT INTO tbl_user(Username, HashedPassword, IsSuper, Mark) VALUES(?,?,?,?);")
	stmt, err := tx.Prepare(sql)
	if err != nil {
		log.Error(err)
		return err
	}
	_, err = stmt.Exec(user.Username, user.HashedPassword, user.IsSuper, user.Mark)
	tx.Commit()
	if err != nil {
		log.Error(err)
		return err
	}
	return nil
}
func QueryAdminUsers() ([]structs.User, error) {
	sql := fmt.Sprintf("SELECT Id, Username, IsSuper, Mark, CreatedAt FROM tbl_user where IsSuper=1")
	rows, err := db.Query(sql)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	var users []structs.User
	for rows.Next() {
		user := structs.User{}
		var t time.Time
		err := rows.Scan(&user.Id, &user.Username, &user.IsSuper, &user.Mark, &t)
		user.CreatedAt = t.Local().Format("2006-01-02 15:04:05")
		if err != nil {
			log.Error(err)
		}
		users = append(users, user)
	}
	return users, nil
}

func QueryUserByName(name string) (*structs.User, error) {
	sql := fmt.Sprintf("SELECT Id, Username, IsSuper, Mark, CreatedAt FROM tbl_user where Username=?")
	row := db.QueryRow(sql, name)
	user := structs.User{}
	var t time.Time
	err := row.Scan(&user.Id, &user.Username, &user.IsSuper, &user.Mark, &t)
	if err != nil {
		return nil, nil
	}
	user.CreatedAt = t.Local().Format("2006-01-02 15:04:05")
	return &user, nil
}

func QueryAllUsers() ([]structs.User, error) {
	sql := fmt.Sprintf("SELECT Id, Username, IsSuper, Mark, CreatedAt FROM tbl_user where IsSuper=0")
	rows, err := db.Query(sql)
	defer rows.Close()
	if err != nil {
		log.Error(err)
		return nil, err
	}
	var users []structs.User
	for rows.Next() {
		user := structs.User{}
		var t time.Time
		err := rows.Scan(&user.Id, &user.Username, &user.IsSuper, &user.Mark, &t)
		user.CreatedAt = t.Local().Format("2006-01-02 15:04:05")
		if err != nil {
			log.Error(err)
			continue
		}
		users = append(users, user)
	}
	return users, nil
}

func QueryTableFieldCount(dbName, tblName string) (int, error) {
	sql := fmt.Sprintf("SELECT COUNT(*) FROM information_schema.columns WHERE table_schema='%s' AND table_name='%s';", dbName, tblName)
	row := db.QueryRow(sql)
	var count int
	err := row.Scan(&count)
	if err != nil {
		return 0, nil
	}
	return count, nil
}

func DropDatabase(dbName string) error {
	sql := fmt.Sprintf("DROP DATABASE IF EXISTS %s;", dbName)
	_, err := db.Exec(sql)
	if err != nil {
		return err
	}
	return nil
}

// AddUserOperatorRecord 新增用户登录记录
func AddUserOperatorRecord(record structs.UserOperatorRecord) error {
	tx, _ := db.Begin()
	sql := fmt.Sprintf("INSERT INTO tbl_user_operator(UserId, Ip, Operator) VALUES(?,?,?);")
	stmt, err := tx.Prepare(sql)
	if err != nil {
		log.Error(err)
		return err
	}
	_, err = stmt.Exec(record.UserId, record.Ip, record.Operator)
	tx.Commit()
	if err != nil {
		log.Error(err)
		return err
	}
	return nil
}

//QueryUserOperatorRecords 获取用户的操作记录 login logout 等
func QueryUserOperatorRecords(offset, count int) (int, []structs.UserOperatorRecord, error) {
	sqlCount := fmt.Sprintf("SELECT count(1) FROM tbl_user_operator")
	ret := db.QueryRow(sqlCount)
	var total int
	err := ret.Scan(&total)
	if err != nil {
		log.Error(err)
		return 0, nil, err
	}

	sql := fmt.Sprintf("SELECT a.Id, a.UserId, b.Username, a.Ip, a.Operator, a.CreateAt FROM tbl_user_operator a INNER JOIN tbl_user b ON a.UserId = b.Id ORDER BY a.CreateAt DESC LIMIT %d,%d", offset, count)
	rows, err := db.Query(sql)
	defer rows.Close()
	if err != nil {
		log.Error(err)
		return 0, nil, err
	}
	var records []structs.UserOperatorRecord
	for rows.Next() {
		record := structs.UserOperatorRecord{}
		var t time.Time
		err := rows.Scan(&record.Id, &record.UserId, &record.Username, &record.Ip, &record.Operator, &t)
		record.CreateAt = t.Local().Format("2006-01-02 15:04:05")
		if err != nil {
			log.Error(err)
			continue
		}
		records = append(records, record)
	}
	return total, records, nil
}

// DeleteUserOperatorRecord 删除操作记录
func DeleteUserOperatorRecord(Id int64) error {
	tx, _ := db.Begin()
	sql := fmt.Sprintf("Delete From tbl_user_operator where Id=?;")
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

//DeleteAllUserOperatorRecords 清理表中数据
func DeleteAllUserOperatorRecords() error {
	tx, _ := db.Begin()
	sql := fmt.Sprintf("Delete From tbl_user_operator;")
	stmt, err := tx.Prepare(sql)
	if err != nil {
		log.Error(err)
		return err
	}
	_, err = stmt.Exec()
	tx.Commit()
	if err != nil {
		log.Error(err)
		return err
	}
	return nil
}

func QueryUserFavorById(userId int64) (string, error) {
	sqlCount := fmt.Sprintf("SELECT Favor FROM tbl_user_favor WHERE UserId=?")
	ret := db.QueryRow(sqlCount, userId)
	var favor string
	err := ret.Scan(&favor)
	if err != nil {
		log.Error(err)
		return "", nil
	}
	return favor, nil
}

func InsertOrReplaceUserFavor(userId int64, newFavor string) error {
	tx, _ := db.Begin()
	sql := fmt.Sprintf("REPLACE INTO tbl_user_favor(UserId, Favor) VALUES(?,?);")
	stmt, err := tx.Prepare(sql)
	if err != nil {
		log.Error(err)
		return err
	}
	_, err = stmt.Exec(userId, newFavor)
	tx.Commit()
	if err != nil {
		log.Error(err)
		return err
	}
	return nil
}
