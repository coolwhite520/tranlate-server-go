package services

import (
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/mvc"
	log "github.com/sirupsen/logrus"
	"strings"
	"translate-server/datamodels"
	"translate-server/docker"
	"translate-server/structs"
)


type ActivationService interface {
	Activation(ctx iris.Context) mvc.Result
}

func NewActivationService() ActivationService {
	return &activationService{}
}

type activationService struct {

}

func (a *activationService) Activation(ctx iris.Context) mvc.Result {
	var jsonObj struct{
		Sn string `json:"sn"`
		Keystore string `json:"keystore"`
	}
	err := ctx.ReadJSON(&jsonObj)
	if err  != nil{
		return mvc.Response{
			Object: map[string]interface{} {
				"code": structs.HttpJsonParseError,
				"msg": structs.HttpJsonParseError.String(),
			},
		}
	}
	jsonObj.Keystore = strings.Trim(jsonObj.Keystore, " ")
	model := datamodels.NewActivationModel()
	activationInfo, state := model.ParseKeystoreContent(jsonObj.Keystore)
	if state != structs.HttpSuccess {
		return mvc.Response{
			Object: map[string]interface{} {
				"code": state,
				"msg": state.String(),
			},
		}
	}
	state = model.GenerateKeystoreFileByContent(jsonObj.Keystore)
	if state != structs.HttpSuccess {
		return mvc.Response{
			Object: map[string]interface{} {
				"code": state,
				"msg": state.String(),
			},
		}
	}
	// 判断是否需要重新写入一个新的过期判定文件
	var expired structs.KeystoreExpired
	expired.LeftTimeSpan = activationInfo.UseTimeSpan
	expired.Sn = activationInfo.Sn
	expired.CreatedAt = activationInfo.CreatedAt

	expiredInfo, state := model.ParseExpiredFile()
	if state == structs.HttpActivationNotFound {
		state = model.GenerateExpiredFile(expired)
		return mvc.Response{
			Object: map[string]interface{} {
				"code": state,
				"msg": state.String(),
			},
		}
	} else {
		// 当时间不同的时候, 需要替换授权
		if expiredInfo.CreatedAt != activationInfo.CreatedAt {
			model.GenerateExpiredFile(expired)
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
