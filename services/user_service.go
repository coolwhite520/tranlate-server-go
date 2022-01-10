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
	UpdateUserMark(user datamodels.User) error
    QueryUserByName(name string) (*datamodels.User, error)
	QueryTableFieldCount(dbName, tblName string) (int, error)
	DropDatabase(dbName string) error

	AddUserOperatorRecord(record datamodels.UserOperatorRecord) error
	QueryUserOperatorRecords(offset, count int) (int, []datamodels.UserOperatorRecord, error)
	DeleteUserOperatorRecord(Id int64) error
	DeleteAllUserOperatorRecords() error
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
	err := row.Scan(&user.Id, &user.Username, &user.HashedPassword, &user.IsSuper, &user.Mark, &user.CreatedAt)
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

func (u *userService) UpdateUserMark(user datamodels.User) error {
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

func (u *userService) InsertUser(user datamodels.User) error {
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
func (u *userService) QueryAdminUsers() ([]datamodels.User, error) {
	sql := fmt.Sprintf("SELECT Id, Username, IsSuper, Mark, CreatedAt FROM tbl_user where IsSuper=1")
	rows, err := db.Query(sql)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	var users []datamodels.User
	for rows.Next() {
		user := datamodels.User{}
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

func (u *userService) QueryUserByName(name string) (*datamodels.User, error) {
	sql := fmt.Sprintf("SELECT Id, Username, IsSuper, Mark, CreatedAt FROM tbl_user where Username=?")
	row := db.QueryRow(sql, name)
	user := datamodels.User{}
	var t time.Time
	err := row.Scan(&user.Id, &user.Username, &user.IsSuper, &user.Mark, &t)
	if err != nil {
		return nil, nil
	}
	user.CreatedAt = t.Local().Format("2006-01-02 15:04:05")
	return &user, nil
}

func (u *userService) QueryAllUsers() ([]datamodels.User, error) {
	sql := fmt.Sprintf("SELECT Id, Username, IsSuper, Mark, CreatedAt FROM tbl_user where IsSuper=0")
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

func (u *userService) QueryTableFieldCount(dbName, tblName string) (int, error) {
	sql := fmt.Sprintf("SELECT COUNT(*) FROM information_schema.columns WHERE table_schema='%s' AND table_name='%s';", dbName, tblName)
	row := db.QueryRow(sql)
	var count int
	err := row.Scan(&count)
	if err != nil {
		return 0, nil
	}
	return count, nil
}

func (u *userService) DropDatabase(dbName string) error {
	sql := fmt.Sprintf("DROP DATABASE IF EXISTS %s;", dbName)
	_, err := db.Exec(sql)
	if err != nil {
		return err
	}
	return nil
}

// AddUserOperatorRecord 新增用户登录记录
func (u *userService) AddUserOperatorRecord(record datamodels.UserOperatorRecord) error {
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
func (u *userService) QueryUserOperatorRecords(offset, count int) (int, []datamodels.UserOperatorRecord, error) {
	sqlCount := fmt.Sprintf("SELECT count(1) FROM tbl_user_operator")
	ret := db.QueryRow(sqlCount)
	var total int
	err := ret.Scan(&total)
	if err != nil {
		log.Error(err)
		return 0, nil, err
	}

	sql := fmt.Sprintf("SELECT a.Id, a.UserId, b.Username, a.Ip, a.Operator, a.CreateAt FROM tbl_user_operator a left join tbl_user b on a.UserId = b.Id order by a.CreateAt DESC limit %d,%d", offset, count)
	rows, err := db.Query(sql)
	defer rows.Close()
	if err != nil {
		log.Error(err)
		return 0, nil, err
	}
	var records []datamodels.UserOperatorRecord
	for rows.Next() {
		record := datamodels.UserOperatorRecord{}
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
func (u *userService) DeleteUserOperatorRecord(Id int64) error {
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
func (u *userService) DeleteAllUserOperatorRecords() error {
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