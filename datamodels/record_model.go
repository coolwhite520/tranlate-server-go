package datamodels

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"strings"
	"time"
	"translate-server/structs"
)

var RecordTableFieldList = []string{
	"Sha1",
	"Content",
	"ContentType",
	"TransType",
	"OutputContent",
	"SrcLang",
	"DesLang",
	"FileName",
	"FileExt",
	"DirRandId",
	"Progress",
	"State",
	"StateDescribe",
	"Error",
	"UserId",
	"OutFileExt",
	"StartAt",
	"EndAt",
}

func DeleteTranslateRecordByIdAndUserId(id int64, userId int64) error {
	tx, _ := db.Begin()
	sql := fmt.Sprintf("DELETE FROM tbl_record where Id=? and UserId=?")
	stmt, err := tx.Prepare(sql)
	if err != nil {
		log.Error(err)
		return err
	}
	_, err = stmt.Exec(id, userId)
	tx.Commit()
	if err != nil {
		log.Error(err)
		return err
	}
	return nil
}

func DeleteTranslateRecordById(id int64) error {
	tx, _ := db.Begin()
	sql := fmt.Sprintf("DELETE FROM tbl_record where Id=?")
	stmt, err := tx.Prepare(sql)
	if err != nil {
		log.Error(err)
		return err
	}
	_, err = stmt.Exec(id)
	tx.Commit()
	if err != nil {
		log.Error(err)
		return err
	}
	return nil
}

// QueryTranslateRecordsBySha1 根据sha1字符串查找数据
func QueryTranslateRecordsBySha1(sha1str string) ([]structs.Record, error) {
	sql := fmt.Sprintf("SELECT * FROM tbl_record where Sha1=?;")
	rows, err := db.Query(sql, sha1str)
	if err != nil {
		return nil, err
	}
	var records []structs.Record
	for rows.Next() {
		var record structs.Record
		var tt time.Time
		err := rows.Scan(
			&record.Id,
			&record.Sha1,
			&record.Content,
			&record.ContentType,
			&record.TransType,
			&record.OutputContent,
			&record.SrcLang,
			&record.DesLang,
			&record.FileName,
			&record.FileExt,
			&record.DirRandId,
			&record.Progress,
			&record.State,
			&record.StateDescribe,
			&record.Error,
			&record.UserId,
			&record.OutFileExt,
			&record.StartAt,
			&record.EndAt,
			&tt)
		if err != nil {
			return nil, err
		}
		record.CreateAt = tt.Local().Format("2006-01-02 15:04:05")
		records = append(records, record)
	}
	return records, nil
}
func QueryTranslateRecordById(id int64) (*structs.Record, error) {
	sql := fmt.Sprintf("SELECT * FROM tbl_record where Id=?;")
	row := db.QueryRow(sql, id)
	record := new(structs.Record)
	var tt time.Time
	err := row.Scan(
		&record.Id,
		&record.Sha1,
		&record.Content,
		&record.ContentType,
		&record.TransType,
		&record.OutputContent,
		&record.SrcLang,
		&record.DesLang,
		&record.FileName,
		&record.FileExt,
		&record.DirRandId,
		&record.Progress,
		&record.State,
		&record.StateDescribe,
		&record.Error,
		&record.UserId,
		&record.OutFileExt,
		&record.StartAt,
		&record.EndAt,
		&tt)
	if err != nil {
		return nil, err
	}
	record.CreateAt = tt.Local().Format("2006-01-02 15:04:05")
	return record, nil
}
func QueryTranslateRecordByIdAndUserId(id int64, userId int64) (*structs.Record, error) {
	sql := fmt.Sprintf("SELECT * FROM tbl_record where Id=? and UserId=?;")
	row := db.QueryRow(sql, id, userId)
	record := new(structs.Record)
	var tt time.Time
	err := row.Scan(
		&record.Id,
		&record.Sha1,
		&record.Content,
		&record.ContentType,
		&record.TransType,
		&record.OutputContent,
		&record.SrcLang,
		&record.DesLang,
		&record.FileName,
		&record.FileExt,
		&record.DirRandId,
		&record.Progress,
		&record.State,
		&record.StateDescribe,
		&record.Error,
		&record.UserId,
		&record.OutFileExt,
		&record.StartAt,
		&record.EndAt,
		&tt)
	if err != nil {
		return nil, err
	}
	record.CreateAt = tt.Local().Format("2006-01-02 15:04:05")
	return record, nil
}
func QueryTranslateRecordsByUserIdAndType(userId int64, transType int, offset int, count int) (int, []structs.Record, error) {
	sqlCount := fmt.Sprintf("SELECT count(1) FROM tbl_record where UserId=? and TransType=?")
	ret := db.QueryRow(sqlCount, userId, transType)
	var total int
	err := ret.Scan(&total)
	if err != nil {
		log.Error(err)
		return 0, nil, err
	}
	sql := fmt.Sprintf("SELECT * FROM tbl_record where UserId=? and TransType=? order by Id DESC limit %d,%d", offset, count)
	rows, err := db.Query(sql, userId, transType)
	if err != nil {
		log.Error(err)
		return 0, nil, err
	}
	var records []structs.Record
	for rows.Next() {
		record := structs.Record{}
		var tt time.Time
		err = rows.Scan(
			&record.Id,
			&record.Sha1,
			&record.Content,
			&record.ContentType,
			&record.TransType,
			&record.OutputContent,
			&record.SrcLang,
			&record.DesLang,
			&record.FileName,
			&record.FileExt,
			&record.DirRandId,
			&record.Progress,
			&record.State,
			&record.StateDescribe,
			&record.Error,
			&record.UserId,
			&record.OutFileExt,
			&record.StartAt,
			&record.EndAt,
			&tt)
		if err != nil {
			return 0, nil, err
		}
		record.CreateAt = tt.Local().Format("2006-01-02 15:04:05")
		records = append(records, record)
	}
	return total, records, nil
}

func QueryTranslateFileRecordsByUserId(userId int64, offset int, count int) (int, []structs.Record, error) {
	sqlCount := fmt.Sprintf("SELECT count(1) FROM tbl_record where UserId=? and TransType != 0")
	ret := db.QueryRow(sqlCount, userId)
	var total int
	err := ret.Scan(&total)
	if err != nil {
		log.Error(err)
		return 0, nil, err
	}

	sql := fmt.Sprintf("SELECT * FROM tbl_record where UserId=? and TransType != 0 order by Id DESC limit %d,%d", offset, count)
	rows, err := db.Query(sql, userId)
	if err != nil {
		log.Error(err)
		return 0, nil, err
	}
	var records []structs.Record
	for rows.Next() {
		record := structs.Record{}
		var tt time.Time
		err = rows.Scan(
			&record.Id,
			&record.Sha1,
			&record.Content,
			&record.ContentType,
			&record.TransType,
			&record.OutputContent,
			&record.SrcLang,
			&record.DesLang,
			&record.FileName,
			&record.FileExt,
			&record.DirRandId,
			&record.Progress,
			&record.State,
			&record.StateDescribe,
			&record.Error,
			&record.UserId,
			&record.OutFileExt,
			&record.StartAt,
			&record.EndAt,
			&tt)
		if err != nil {
			return 0, nil, err
		}
		record.CreateAt = tt.Local().Format("2006-01-02 15:04:05")
		records = append(records, record)
	}
	return total, records, nil
}

func QueryTranslateRecords(offset int, count int) (int, []structs.RecordEx, error) {
	sqlCount := fmt.Sprintf("SELECT count(1)  FROM tbl_record a INNER JOIN tbl_user b ON a.UserId = b.Id;")
	ret := db.QueryRow(sqlCount)
	var total int
	err := ret.Scan(&total)
	if err != nil {
		log.Error(err)
		return 0, nil, err
	}
	sql := fmt.Sprintf("SELECT a.*, b.Username as Username  FROM tbl_record a INNER JOIN tbl_user b ON a.UserId = b.Id order by Id DESC limit %d,%d", offset, count)
	rows, err := db.Query(sql)
	if err != nil {
		log.Error(err)
		return 0, nil, err
	}
	var records []structs.RecordEx
	for rows.Next() {
		record := structs.RecordEx{}
		var tt time.Time
		err = rows.Scan(
			&record.Id,
			&record.Sha1,
			&record.Content,
			&record.ContentType,
			&record.TransType,
			&record.OutputContent,
			&record.SrcLang,
			&record.DesLang,
			&record.FileName,
			&record.FileExt,
			&record.DirRandId,
			&record.Progress,
			&record.State,
			&record.StateDescribe,
			&record.Error,
			&record.UserId,
			&record.OutFileExt,
			&record.StartAt,
			&record.EndAt,
			&tt,
			&record.UserName,
		)
		if err != nil {
			return 0, nil, err
		}
		record.CreateAt = tt.Local().Format("2006-01-02 15:04:05")
		records = append(records, record)
	}
	return total, records, nil
}

func QueryTranslateRecordsByUserId(userId int64) ([]structs.Record, error) {
	sql := fmt.Sprintf("SELECT * FROM tbl_record where UserId=? order by Id DESC")
	rows, err := db.Query(sql, userId)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	var records []structs.Record
	for rows.Next() {
		record := structs.Record{}
		var tt time.Time
		err = rows.Scan(
			&record.Id,
			&record.Sha1,
			&record.Content,
			&record.ContentType,
			&record.TransType,
			&record.OutputContent,
			&record.SrcLang,
			&record.DesLang,
			&record.FileName,
			&record.FileExt,
			&record.DirRandId,
			&record.Progress,
			&record.State,
			&record.StateDescribe,
			&record.Error,
			&record.UserId,
			&record.OutFileExt,
			&record.StartAt,
			&record.EndAt,
			&tt)
		if err != nil {
			return nil, err
		}
		record.CreateAt = tt.Local().Format("2006-01-02 15:04:05")
		records = append(records, record)
	}
	return records, nil
}


func UpdateRecordProgress(Id int64, Progress int) {
	tx, _ := db.Begin()
	sql := fmt.Sprintf("UPDATE tbl_record set Progress=? where Id=?;")
	stmt, err := tx.Prepare(sql)
	if err != nil {
		log.Error(err)
		return
	}
	_, err = stmt.Exec(
		Progress,
		Id,
	)
	tx.Commit()
	if err != nil {
		log.Error(err)
		return
	}
	return
}

func UpdateRecord(record *structs.Record)  {
	var q []string
	for _, _ = range RecordTableFieldList {
		q = append(q, "?")
	}
	allFields := strings.Join(RecordTableFieldList, ",")
	allQs := strings.Join(q, ",")
	tx, _ := db.Begin()
	sql := fmt.Sprintf("REPLACE INTO tbl_record(Id,%s) VALUES(?,%s);", allFields, allQs)
	stmt, err := tx.Prepare(sql)
	if err != nil {
		log.Error(err)
		return
	}
	_, err = stmt.Exec(
		record.Id,
		record.Sha1,
		record.Content,
		record.ContentType,
		record.TransType,
		record.OutputContent,
		record.SrcLang,
		record.DesLang,
		record.FileName,
		record.FileExt,
		record.DirRandId,
		record.Progress,
		record.State,
		record.StateDescribe,
		record.Error,
		record.UserId,
		record.OutFileExt,
		record.StartAt,
		record.EndAt,
	)
	tx.Commit()
	if err != nil {
		log.Error(err)
		return
	}
	return
}

func InsertRecord(record *structs.Record) error {
	var q []string
	for _, _ = range RecordTableFieldList {
		q = append(q, "?")
	}
	allFields := strings.Join(RecordTableFieldList, ",")
	allQs := strings.Join(q, ",")
	tx, _ := db.Begin()
	sql := fmt.Sprintf("INSERT INTO tbl_record(%s) VALUES(%s);", allFields, allQs)
	stmt, err := tx.Prepare(sql)
	if err != nil {
		log.Error(err)
		return err
	}
	result, err := stmt.Exec(
		record.Sha1,
		record.Content,
		record.ContentType,
		record.TransType,
		record.OutputContent,
		record.SrcLang,
		record.DesLang,
		record.FileName,
		record.FileExt,
		record.DirRandId,
		record.Progress,
		record.State,
		record.StateDescribe,
		record.Error,
		record.UserId,
		record.OutFileExt,
		record.StartAt,
		record.EndAt,
	)
	tx.Commit()
	if err != nil {
		log.Error(err)
		return err
	}
	record.Id, _ = result.LastInsertId()
	return nil
}
