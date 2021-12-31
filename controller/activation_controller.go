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
	NewActivation services.ActivationService
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
	activationInfo, state := a.NewActivation.ParseKeystoreContent(jsonObj.Keystore)
	if state != datamodels.HttpSuccess {
		return mvc.Response{
			Object: map[string]interface{} {
				"code": state,
				"msg": state.String(),
			},
		}
	}
	state = a.NewActivation.GenerateKeystoreFileByContent(jsonObj.Keystore)
	if state != datamodels.HttpSuccess {
		return mvc.Response{
			Object: map[string]interface{} {
				"code": state,
				"msg": state.String(),
			},
		}
	}
	// 判断是否需要重新写入一个新的过期判定文件
	var expired datamodels.KeystoreExpired
	expired.LeftTimeSpan = activationInfo.UseTimeSpan
	expired.Sn = activationInfo.Sn
	expired.CreatedAt = activationInfo.CreatedAt

	expiredInfo, state := a.NewActivation.ParseExpiredFile()
	if state == datamodels.HttpActivationNotFound {
		state = a.NewActivation.GenerateExpiredFile(expired)
		return mvc.Response{
			Object: map[string]interface{} {
				"code": state,
				"msg": state.String(),
			},
		}
	} else {
		// 当时间不同的时候, 需要替换授权
		if expiredInfo.CreatedAt != activationInfo.CreatedAt {
			a.NewActivation.GenerateExpiredFile(expired)
		}
	}
	go func() {
		if docker.GetInstance().GetStatus() == docker.InitializingStatus {
			return
		}
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


