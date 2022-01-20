package translate_models

import (
	"fmt"
	"github.com/kataras/iris/v12"
	"os"
	"path/filepath"
	"strings"
	"time"
	"translate-server/datamodels"
	"translate-server/structs"
	"translate-server/utils"
)

func ReceiveFiles(Ctx iris.Context) ([]structs.Record, error) {
	var records []structs.Record
	u := Ctx.Values().Get("User")
	user, _ := (u).(structs.User)
	// 创建用户的子目录
	nowUnixMicro := time.Now().UnixMicro()
	DirRandId := fmt.Sprintf("%d", nowUnixMicro)
	userUploadDir := fmt.Sprintf("%s/%d/%s", structs.UploadDir, user.Id, DirRandId)
	if !utils.PathExists(userUploadDir) {
		err := os.MkdirAll(userUploadDir, os.ModePerm)
		if err != nil {
			return records, err
		}
	}
	files, _, err := Ctx.UploadFormFiles(userUploadDir)
	if err != nil {
		return records, err
	}
	for _, v := range files {
		filePath := fmt.Sprintf("%s/%s", userUploadDir, v.Filename)
		contentType, _ := utils.GetFileContentType(filePath)
		fileExt := filepath.Ext(v.Filename)
		fileName := v.Filename[0:len(v.Filename) - len(fileExt)]
		var TransType int
		var OutFileExt string
		if strings.Contains(contentType, "image/") {
			TransType = 1
			OutFileExt = ".docx"
		} else {
			TransType = 2
			OutFileExt = fileExt
		}

		record := structs.Record{
			ContentType:   contentType,
			TransType:     TransType,
			FileName:      fileName,
			FileExt:       fileExt,
			DirRandId:     DirRandId,
			CreateAt:      time.Now().Format("2006-01-02 15:04:05"),
			State:         structs.TransNoRun,
			StateDescribe: structs.TransNoRun.String(),
			UserId:        user.Id,
			OutFileExt:    OutFileExt,
		}
		err = datamodels.InsertRecord(&record)
		if err != nil {
			continue
		}
		records = append(records, record)
	}
	return records, nil
}


