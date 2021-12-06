package controller

import (
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/mvc"
	"translate-server/activation"
)

type ActivationController struct {
	Ctx iris.Context
}

func (a *ActivationController) BeforeActivation(b mvc.BeforeActivation) {

}

func (a *ActivationController) Post() mvc.Result {
	var jsonObj struct{
		MachineId string `json:"machine_id"`
		Keystore string `json:"keystore"`
	}
	err := a.Ctx.ReadJSON(&jsonObj)
	if err  != nil{
		return mvc.Response{
			Err: err,
		}
	}
	newActivation := activation.NewActivation()
	aInfo, state := newActivation.ParseKeystoreContent(jsonObj.Keystore)
	if state != activation.Success {
		return mvc.Response{
			Text: state.String(),
		}
	}
	state = newActivation.GenerateKeystoreFile(*aInfo)
	if state != activation.Success {
		return mvc.Response{
			Text: state.String(),
		}
	}
	return mvc.Response{
		Text: state.String(),
	}
}

