package datamodels

import (
"fmt"
"github.com/Unknwon/goconfig"
log "github.com/sirupsen/logrus"
"time"
"translate-server/structs"
)

func  GetIpTableType() (string, error) {
	cfg, err := goconfig.LoadConfigFile("./config.ini")
	if err != nil {
		return "", err
	}
	value, err := cfg.GetValue("IP_TABLE", "type")
	if err != nil {
		return "", err
	}
	return value, nil
}

func  SetIpTableType(tType string) (bool, error) {
	cfg, err := goconfig.LoadConfigFile("./config.ini")
	if err != nil {
		return false, err
	}
	cfg.SetValue("IP_TABLE", "type", tType)
	err = goconfig.SaveConfigFile(cfg, "./config.ini")
	if err != nil {
		return false, err
	}
	return true, nil
}

func  AddIpTblRecord(record structs.IpTableRecord) error {
	tx, _ := db.Begin()
	sql := fmt.Sprintf("INSERT INTO tbl_ips(Ip, Type) VALUES(?,?);")
	stmt, err := tx.Prepare(sql)
	if err != nil {
		log.Error(err)
		return err
	}
	_, err = stmt.Exec(record.Ip, record.Type)
	tx.Commit()
	if err != nil {
		log.Error(err)
		return err
	}
	return nil
}

func DelIpTblRecord(Id int64) error {
	tx, _ := db.Begin()
	sql := fmt.Sprintf("DELETE FROM tbl_ips WHERE Id=?;")
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

func QueryIpTblRecords() ([]structs.IpTableRecord, error) {
	sql := fmt.Sprintf("SELECT Id, Ip, Type, CreateAt FROM tbl_ips ORDER BY CreateAt DESC;")
	rows, err := db.Query(sql)
	defer rows.Close()
	if err != nil {
		log.Error(err)
		return nil, err
	}
	var records []structs.IpTableRecord
	for rows.Next() {
		record := structs.IpTableRecord{}
		var t time.Time
		err := rows.Scan(&record.Id, &record.Ip, &record.Type, &t)
		record.CreateAt = t.Local().Format("2006-01-02 15:04:05")
		if err != nil {
			log.Error(err)
			continue
		}
		records = append(records, record)
	}
	return records, nil
}