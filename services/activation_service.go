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
	bannedList, _ := model.ParseBannedFile()
	if expiredInfo == nil {
		// 1 未激活的判定
		proof.Sn = model.GetMachineId()
		proof.State = structs.ProofStateNotActivation
		proof.StateDescribe = proof.State.String()
		proof.Id = 0
	} else if bannedList != nil {
		// 2 过期和 失效的判断
		for _, v := range bannedList {
			if v.Id == expiredInfo.CreatedAt && v.State == structs.ProofStateForceBanned{
				proof.Sn = model.GetMachineId()
				proof.State = structs.ProofStateForceBanned
				proof.StateDescribe = proof.State.String()
				proof.Id = expiredInfo.CreatedAt
				break
			}
			if v.Id == expiredInfo.CreatedAt && v.State == structs.ProofStateExpired{
				proof.Sn = model.GetMachineId()
				proof.State = structs.ProofStateExpired
				proof.StateDescribe = proof.State.String()
				proof.Id = expiredInfo.CreatedAt
				break
			}
		}
	} else {
		proof.Sn = model.GetMachineId()
		proof.State = structs.ProofStateOk
		proof.StateDescribe = proof.State.String()
		proof.Id = expiredInfo.CreatedAt
	}

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
	var banInfo structs.BannedInfo
	banInfo.Id = activationInfo.CreatedAt
	banInfo.State = structs.ProofStateForceBanned
	banInfo.StateDescribe = banInfo.State.String()
	code := model.AddId2BannedFile(banInfo)
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
	bannedList, _ := model.ParseBannedFile()
	if bannedList != nil {
		for _, v := range bannedList {
			if v.Id == activationInfo.CreatedAt && v.State == structs.ProofStateForceBanned {
				return mvc.Response{
					Object: map[string]interface{} {
						"code": constant.HttpActivationInvalidateError,
						"msg":  "输入了已失效的授权。",
					},
				}
			}
			if v.Id == activationInfo.CreatedAt && v.State == structs.ProofStateExpired {
				return mvc.Response{
					Object: map[string]interface{}{
						"code": constant.HttpActivationExpiredError,
						"msg":  "输入了已过期的授权。",
					},
				}
			}
		}
	}
	// 判断是否存在老的授权
	oldKeystore, _ := model.ParseKeystoreFile()
	if oldKeystore != nil {
		if oldKeystore.CreatedAt == activationInfo.CreatedAt {
			return mvc.Response {
				Object: map[string]interface{}{
					"code": constant.HttpActivationInvalidateError,
					"msg":  "无法导入正在使用中的授权。",
				},
			}
		}
		// 将老的授权CreateAt加入到ban列表中
		var bannedInfo structs.BannedInfo
		bannedInfo.Id = oldKeystore.CreatedAt
		bannedInfo.State = structs.ProofStateForceBanned
		bannedInfo.StateDescribe = bannedInfo.State.String()
		model.AddId2BannedFile(bannedInfo)
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
	nowTs := time.Now().Unix()
	expired.ActivationAt = nowTs
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
