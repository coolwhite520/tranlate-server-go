package translate_models

import (
	"baliance.com/gooxml/document"
	"fmt"
	log "github.com/sirupsen/logrus"
	"os"
	"path/filepath"
	"strings"
	"translate-server/apis"
	"translate-server/datamodels"
	"translate-server/structs"
	"translate-server/utils"
)

func calculateDocTotalProgress(srcFilePathName string) (int, error) {
	sum := 0
	doc, err := document.Open(srcFilePathName)
	if err != nil {
		log.Errorln(err)
		return 0, err
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
			sum ++
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
				sum++
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
						sum++
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
				sum ++
			}
		}
	}
	return sum, nil
}

func translateDocxFile(srcLang string, desLang string, record *structs.Record) error {
	srcDir := fmt.Sprintf("%s/%d/%s", structs.UploadDir, record.UserId, record.DirRandId)
	translatedDir := fmt.Sprintf("%s/%d/%s", structs.OutputDir, record.UserId, record.DirRandId)
	srcFilePathName := fmt.Sprintf("%s/%s%s", srcDir, record.FileName, record.FileExt)
	ext := filepath.Ext(record.FileExt)
	if strings.ToLower(ext) == ".doc" {
		err := apis.PyConvertSpecialFile(srcFilePathName, "d2dx")
		if err != nil {
			return err
		}
		srcFilePathName = srcFilePathName + "x"
	}
	totalProgress, err := calculateDocTotalProgress(srcFilePathName)
	if err != nil {
		return err
	}
	currentProgress := 0
	percent := 0
	doc, err := document.Open(srcFilePathName)
	if err != nil {
		return err
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
			currentProgress++
			if percent != currentProgress * 100 /totalProgress{
				percent = currentProgress * 100 /totalProgress
				datamodels.UpdateRecordProgress(record.Id, percent)
			}
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
				currentProgress++
				if percent != currentProgress * 100 /totalProgress{
					percent = currentProgress * 100 /totalProgress
					datamodels.UpdateRecordProgress(record.Id, percent)
				}
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
						currentProgress++
						if percent != currentProgress * 100 /totalProgress{
							percent = currentProgress * 100 /totalProgress
							datamodels.UpdateRecordProgress(record.Id, percent)
						}
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
				currentProgress++
				if percent != currentProgress * 100 /totalProgress{
					percent = currentProgress * 100 /totalProgress
					datamodels.UpdateRecordProgress(record.Id, percent)
				}
				run.AddText(transContent)
			}
		}

	}
	if !utils.PathExists(translatedDir) {
		err := os.MkdirAll(translatedDir, os.ModePerm)
		if err != nil {
			return err
		}
	}
	desFile := fmt.Sprintf("%s/%s%s", translatedDir, record.FileName, record.OutFileExt)
	err = doc.SaveToFile(desFile)
	if err != nil {
		return err
	}
	// 计算文件md5
	md5, err := utils.GetFileMd5(srcFilePathName)
	if err != nil {
		return err
	}
	// 拼接sha1字符串
	sha1 := utils.Sha1(fmt.Sprintf("%s&%s&%s", md5, srcLang, desLang))
	record.Sha1 = sha1
	datamodels.UpdateRecord(record)
	return nil
}
