package controller

import (
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/mvc"
	"strings"
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
			Object: map[string]interface{} {
				"code": -100,
				"msg": err.Error(),
			},
		}
	}
	jsonObj.Keystore = strings.Trim(jsonObj.Keystore, " ")
	newActivation := activation.NewActivation()

	_, state := newActivation.ParseKeystoreContent(jsonObj.Keystore)
	if state != activation.Success {
		return mvc.Response{
			Object: map[string]interface{} {
				"code": -100,
				"msg": state.String(),
			},
		}
	}
	state = newActivation.GenerateKeystoreFileByContent(jsonObj.Keystore)
	if state != activation.Success {
		return mvc.Response{
			Object: map[string]interface{} {
				"code": -100,
				"msg": state.String(),
			},
		}
	}
	return mvc.Response{
		Object: map[string]interface{} {
			"code": 200,
			"msg": state.String(),
		},
	}
}

