package controllers

import (
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/mvc"
	"os"
	"path/filepath"
)

type FileController struct {
	Ctx iris.Context  //上下文对象
}

const maxSize = 80 * iris.MB

func (c *FileController) BeforeActivation(b mvc.BeforeActivation) {
	// b.Dependencies().Add/Remove
	// b.Router().Use/UseGlobal/Done // 和你已知的任何标准 API  调用

	// 1-> 方法
	// 2-> 路径
	// 3-> 控制器函数的名称将被解析未一个处理程序 [ handler ]
	// 4-> 任何应该在 MyCustomHandler 之前运行的处理程序[ handlers ]
	//b.Handle("Post", "/{id:int64}", "MyPost")
}

func (c *FileController) Post() mvc.Result {
	c.Ctx.SetMaxRequestBodySize(maxSize)
	_, fileHeader, err := c.Ctx.FormFile("file")
	if err != nil {
		c.Ctx.StopWithError(iris.StatusBadRequest, err)
		return mvc.Response{
			ContentType: "application/json",
			Text:        err.Error(),
			Code:        -1,
		}
	}
	os.MkdirAll("./uploads", 0777)
	dest := filepath.Join("./uploads", fileHeader.Filename)
	_, err = c.Ctx.SaveFormFile(fileHeader, dest)
	if err != nil {
		c.Ctx.StopWithError(iris.StatusBadRequest, err)
		return mvc.Response{
			ContentType: "application/json",
			Text:        err.Error(),
			Code:        -1,
		}
	}
	c.Ctx.Writef("File: %s uploaded!", fileHeader.Filename)
	return mvc.Response{
		ContentType: "application/json",
		Text:        "success",
		Code:        200,
	}
}
