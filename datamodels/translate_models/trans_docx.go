package translate_models

import (
	"baliance.com/gooxml/document"
	"fmt"
	log "github.com/sirupsen/logrus"
	"os"
	"strings"
	"translate-server/datamodels"
	"translate-server/structs"
	"translate-server/utils"
)

func translateDocxFile(srcLang string, desLang string, record *structs.Record) {
	srcDir := fmt.Sprintf("%s/%d/%s", structs.UploadDir, record.UserId, record.DirRandId)
	translatedDir := fmt.Sprintf("%s/%d/%s", structs.OutputDir, record.UserId, record.DirRandId)
	srcFilePathName := fmt.Sprintf("%s/%s%s", srcDir, record.FileName, record.FileExt)
	record.SrcLang = srcLang
	record.DesLang = desLang
	record.State = structs.TransExtractSuccess
	record.StateDescribe = structs.TransExtractSuccess.String()
	datamodels.UpdateRecord(record)
	doc, err := document.Open(srcFilePathName)
	if err != nil {
		log.Errorln(err)
		return
	}
	paragraphs := doc.Paragraphs()
	for _, p := range paragraphs {
		var content string
		for _, r := range p.Runs() {
			content += r.Text()
		}
		if len(strings.Trim(content, " ")) > 0 {
			for _, r := range p.Runs() {
				p.RemoveRun(r)
			}
			run := p.AddRun()
			transContent, _, _ := translate(srcLang, desLang, content)
			run.AddText(transContent)
		}
	}
	headers := doc.Headers()
	for _, h := range headers {
		for _, p := range h.Paragraphs() {
			var content string
			for _, r := range p.Runs() {
				content += r.Text()
			}
			if len(strings.Trim(content, " ")) > 0 {
				for _, r := range p.Runs() {
					p.RemoveRun(r)
				}
				run := p.AddRun()
				transContent, _, _ := translate(srcLang, desLang, content)
				run.AddText(transContent)
			}
		}
	}
	tables := doc.Tables()
	for _, tal := range tables {
		for _, r := range tal.Rows() {
			for _, c := range r.Cells() {
				for _, p := range c.Paragraphs() {
					var content string
					for _, r := range p.Runs() {
						content += r.Text()
					}
					if len(strings.Trim(content, " ")) > 0 {
						for _, r := range p.Runs() {
							p.RemoveRun(r)
						}
						run := p.AddRun()
						transContent, _, _ := translate(srcLang, desLang, content)
						run.AddText(transContent)
					}
				}

			}
		}
	}

	footers := doc.Footers()
	for _, f := range footers {
		for _, p := range f.Paragraphs() {
			var content string
			for _, r := range p.Runs() {
				content += r.Text()
			}
			if len(strings.Trim(content, " ")) > 0 {
				for _, r := range p.Runs() {
					p.RemoveRun(r)
				}
				run := p.AddRun()
				transContent, _, _ := translate(srcLang, desLang, content)
				run.AddText(transContent)
			}
		}

	}
	if !utils.PathExists(translatedDir) {
		err := os.MkdirAll(translatedDir, os.ModePerm)
		if err != nil {
			return
		}
	}
	desFile := fmt.Sprintf("%s/%s%s", translatedDir, record.FileName, record.OutFileExt)
	err = doc.SaveToFile(desFile)
	if err != nil {
		return
	}
	// 计算文件md5
	md5, err := utils.GetFileMd5(srcFilePathName)
	if err != nil {
		return
	}
	// 拼接sha1字符串
	sha1 := utils.Sha1(fmt.Sprintf("%s&%s&%s", md5, srcLang, desLang))
	record.Sha1 = sha1
	record.State = structs.TransTranslateSuccess
	record.StateDescribe = structs.TransTranslateSuccess.String()
	record.Error = ""
	datamodels.UpdateRecord(record)
}
