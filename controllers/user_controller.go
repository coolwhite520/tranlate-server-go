package controllers

import "github.com/kataras/iris/v12/mvc"

type UserController struct {

}

func (c *UserController) Get() mvc.Result  {
	obj := make(map[string]interface{})
	obj["name"] = "panda"
	obj["sex"] = "man"
	obj["age"] = 10
	return mvc.Response{
		ContentType: "application/json",
		Object: obj,
	}
}