package middleware

import (
	"github.com/kataras/iris/v12"
	"translate-server/datamodels"
	"translate-server/services"
)

func CheckActivationMiddleware(ctx iris.Context) {
	newActivation := services.NewActivationService()
	sn, _ := newActivation.GenerateMachineId()
	//var langList []datamodels.SupportLang
	//langList = append(langList, datamodels.SupportLang{
	//	EnName: "English",
	//	CnName: "英语",
	//}, datamodels.SupportLang{
	//	EnName: "Chinese",
	//	CnName: "中文(简体)",
	//})
	//activationInfo := datamodels.Activation{
	//	UserName:        "panda",
	//	SupportLangList: langList,
	//	CreatedAt:       time.Now().Format("2006-01-02 15:04:05"),
	//	ExpiredAt:       time.Date(2099, 1, 1, 1, 1, 1, 1, time.Local).Format("2006-01-02 15:04:05"),
	//	Sn:       sn,
	//}
	//content, state := newActivation.GenerateKeystoreContent(activationInfo)
	//if state != datamodels.HttpSuccess {
	//
	//}
	_, state := newActivation.ParseKeystoreFile()
	if state != datamodels.HttpSuccess {
		ctx.JSON(
			map[string]interface{}{
				"code":      state,
				"sn":        sn,
				"msg":       state.String(),
				//"keystore":  content,
			})
		return
	}
	ctx.Next()
}
