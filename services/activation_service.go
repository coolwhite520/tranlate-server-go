package services

import (
	"encoding/base64"
	"encoding/json"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/mvc"
	log "github.com/sirupsen/logrus"
	"strings"
	"time"
	"translate-server/constant"
	"translate-server/datamodels"
	"translate-server/docker"
	"translate-server/structs"
	"translate-server/utils"
)

const AesProofKey = "c64df51f4aba41edbaf024cbf73e9234"

type ActivationService interface {
	Activation(ctx iris.Context) mvc.Result
	AddBan() mvc.Result
	PostActivationProof(ctx iris.Context)
}

func NewActivationService() ActivationService {
	return &activationService{}
}

type activationService struct {
}

// PostActivationProof 获取授权凭证状态：0。状态良好 1。未激活 2。过期 3。强制失效
func (a *activationService) PostActivationProof(ctx iris.Context) {
	var proof structs.KeystoreProof
	model := datamodels.NewActivationModel()
	expiredInfo, _ := model.ParseExpiredFile()
	banInfo, _ := model.ParseBannedFile()
	if expiredInfo == nil {
		// 1 未激活的判定
		proof.Sn = model.GetMachineId()
		proof.State = 1
	} else if expiredInfo.LeftTimeSpan <= 0 {
		// 2 过期的判定
		proof.Sn = model.GetMachineId()
		proof.State = 2
	} else if banInfo != nil {
		// 3 强制失效的判定，用户替换授权
		isBaned := false
		for _, v := range banInfo.Ids {
			if v == expiredInfo.CreatedAt {
				isBaned = true
				break
			}
		}
		if isBaned {
			proof.Sn = model.GetMachineId()
			proof.State = 3
		} else {
			proof.Sn = model.GetMachineId()
			proof.State = 0
		}
	} else {
		proof.Sn = model.GetMachineId()
		proof.State = 0
	}
	proof.Now = time.Now().Unix()
	var resultStr string
	bytes, _ := json.Marshal(&proof)
	encrypt, err := utils.AesEncrypt(bytes, []byte(AesProofKey))
	if err != nil {
		resultStr = base64.StdEncoding.EncodeToString(bytes)
	} else {
		resultStr = base64.StdEncoding.EncodeToString(encrypt)
	}
	ctx.ResponseWriter().Write([]byte(resultStr))
}

func (a *activationService) AddBan() mvc.Result {
	model := datamodels.NewActivationModel()
	activationInfo, state := model.ParseKeystoreFile()
	if state != constant.HttpSuccess {
		return mvc.Response{
			Object: map[string]interface{} {
				"code": state,
				"msg": state.String(),
			},
		}
	}
	code := model.AddId2BannedFile(activationInfo.CreatedAt)
	return mvc.Response{
		Object: map[string]interface{} {
			"code": code,
			"msg":  code.String(),
			"data": map[string]interface{} {
				"id": activationInfo.CreatedAt,
				"sn": activationInfo.Sn,
			},
		},
	}
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
				"code": constant.HttpJsonParseError,
				"msg":  constant.HttpJsonParseError.String(),
			},
		}
	}
	jsonObj.Keystore = strings.Trim(jsonObj.Keystore, " ")
	model := datamodels.NewActivationModel()
	activationInfo, state := model.ParseKeystoreContent(jsonObj.Keystore)
	if state != constant.HttpSuccess {
		return mvc.Response{
			Object: map[string]interface{} {
				"code": state,
				"msg": state.String(),
			},
		}
	}
	// 是否永久失效的授权
	bannedInfo, _ := model.ParseBannedFile()
	if bannedInfo != nil {
		for _, v := range bannedInfo.Ids {
			if v == activationInfo.CreatedAt {
				return mvc.Response{
					Object: map[string]interface{} {
						"code": constant.HttpActivationInvalidateError,
						"msg":  "输入了已经失效的授权。",
					},
				}
			}
		}
	}
	// 判断授权是否是已经过期的
	expiredInfo, state := model.ParseExpiredFile()
	if expiredInfo != nil && expiredInfo.CreatedAt == activationInfo.CreatedAt && expiredInfo.LeftTimeSpan <= 0 {
		return mvc.Response {
			Object: map[string]interface{}{
				"code": constant.HttpActivationExpiredError,
				"msg":  constant.HttpActivationExpiredError.String(),
			},
		}
	}
	// 判断是否存在老的授权
	oldKeystore, _ := model.ParseKeystoreFile()
	if oldKeystore != nil {
		if oldKeystore.CreatedAt == activationInfo.CreatedAt {
			return mvc.Response {
				Object: map[string]interface{}{
					"code": constant.HttpActivationInvalidateError,
					"msg":  "导入的授权已存在。",
				},
			}
		}
		// 将老的授权CreateAt加入到ban列表中
		model.AddId2BannedFile(oldKeystore.CreatedAt)
	}

	// 生成授权文件
	state = model.GenerateKeystoreFileByContent(jsonObj.Keystore)
	if state != constant.HttpSuccess {
		return mvc.Response{
			Object: map[string]interface{} {
				"code": state,
				"msg": state.String(),
			},
		}
	}
	// 生成过期文件
	var expired structs.KeystoreExpired
	expired.LeftTimeSpan = activationInfo.UseTimeSpan
	expired.Sn = activationInfo.Sn
	expired.CreatedAt = activationInfo.CreatedAt
	state = model.GenerateExpiredFile(expired)
	if state != constant.HttpSuccess {
		return mvc.Response{
			Object: map[string]interface{} {
				"code": state,
				"msg": state.String(),
			},
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
			"code": constant.HttpSuccess,
			"msg": constant.HttpSuccess.String(),
		},
	}
}
