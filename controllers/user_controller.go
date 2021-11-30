package controllers

import "github.com/kataras/iris/v12/mvc"

type UserController struct {

}

func (u *UserController) Get() mvc.Result  {
	var v struct{
		Name string `json:"name"`
		Age int `json:"age"`
	}
	v.Name = "panda"
	v.Age = 10
	return mvc.Response{
		ContentType: "application/json",
		Object: v,
	}
}