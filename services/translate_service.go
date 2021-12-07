package services

import (
	"net/http"
	"os"
	"strings"
)

type FileType string

const (
	ImageType = "image"
	TextType = "file"
)

type TranslateService interface {
	TranslateContent(srcLang, desLang, content string) (string, error)
	TranslateFile(srcLang, desLang, filePath string) (string, error)
}

func NewTranslateService() TranslateService {
	return &translateService{}
}

type translateService struct {
}

func (t *translateService) getFileContentType(filepath string) (FileType, error) {
	ff, err := os.Open(filepath)
	if err != nil {
		panic(err)
	}
	defer ff.Close()

	buffer := make([]byte, 512)
	_, err = ff.Read(buffer)
	if err != nil {
		return "", err
	}
	contentType := http.DetectContentType(buffer)
	if strings.Contains(contentType, "image/") {
		return ImageType, nil
	}
	return TextType, nil
}

func (t translateService) ocrDetectedImage(filePath string) (string, error) {
	return "", nil
}

func (t translateService) tikaDetectedText(filePath string) (string, error) {
	return "", nil
}

func  (t * translateService) extractContent(filePath string) (string, error) {
	contentType, err := t.getFileContentType(filePath)
	if err != nil {
		return "", err
	}
	if contentType == TextType {
		return t.tikaDetectedText(filePath)
	} else {
		return t.ocrDetectedImage(filePath)
	}
}

func (t * translateService) TranslateContent(srcLang, desLang, content string) (string, error){
	return "",nil
}

func (t * translateService) TranslateFile(srcLang, desLang, filePath string) (string, error) {
	content, err := t.extractContent(filePath)
	if err != nil {
		return "", err
	}
	return t.TranslateContent(srcLang, desLang, content)
}