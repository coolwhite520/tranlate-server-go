package server

import (
	"github.com/kataras/iris/v12"
	"os"
	"path/filepath"
)

const maxSize = 80 * iris.MB

func uploadFile(ctx iris.Context) {
	// Set a lower memory limit for multipart forms (default is 32 MiB)
	ctx.SetMaxRequestBodySize(maxSize)
	// single file
	_, fileHeader, err:= ctx.FormFile("file")
	if err != nil {
		ctx.StopWithError(iris.StatusBadRequest, err)
		return
	}
	// Upload the file to specific destination.
	os.MkdirAll("./uploads", 0777)
	dest := filepath.Join("./uploads", fileHeader.Filename)
	_, err = ctx.SaveFormFile(fileHeader, dest)
	if err != nil {
		ctx.StopWithError(iris.StatusBadRequest, err)
		return
	}
	ctx.Writef("File: %s uploaded!", fileHeader.Filename)
}

