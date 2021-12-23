package controller

import (
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/mvc"
	log "github.com/sirupsen/logrus"
	"strings"
	"translate-server/datamodels"
	"translate-server/docker"
	"translate-server/services"
)

type ActivationController struct {
	Ctx iris.Context
}

func (a *ActivationController) BeforeActivation(b mvc.BeforeActivation) {

}

func (a *ActivationController) Post() mvc.Result {
	var jsonObj struct{
		Sn string `json:"sn"`
		Keystore string `json:"keystore"`
	}
	err := a.Ctx.ReadJSON(&jsonObj)
	if err  != nil{
		return mvc.Response{
			Object: map[string]interface{} {
				"code": datamodels.HttpJsonParseError,
				"msg": datamodels.HttpJsonParseError.String(),
			},
		}
	}
	jsonObj.Keystore = strings.Trim(jsonObj.Keystore, " ")
	newActivation := services.NewActivationService()

	_, state := newActivation.ParseKeystoreContent(jsonObj.Keystore)
	if state != datamodels.HttpSuccess {
		return mvc.Response{
			Object: map[string]interface{} {
				"code": state,
				"msg": state.String(),
			},
		}
	}
	state = newActivation.GenerateKeystoreFileByContent(jsonObj.Keystore)
	if state != datamodels.HttpSuccess {
		return mvc.Response{
			Object: map[string]interface{} {
				"code": state,
				"msg": state.String(),
			},
		}
	}
	go func() {
		docker.GetInstance().SetStatus(docker.InitializingStatus)
		err = docker.GetInstance().StartDockers()
		if err != nil {
			log.Error(err)
			return
		}
		docker.GetInstance().SetStatus(docker.NormalStatus)
	}()
	return mvc.Response{
		Object: map[string]interface{} {
			"code": state,
			"msg": state.String(),
		},
	}
}


