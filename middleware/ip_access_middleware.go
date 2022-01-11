package middleware

import (
	"github.com/kataras/iris/v12"
	"github.com/thinkeridea/go-extend/exnet"
	"log"
	"strings"
	"translate-server/datamodels"
	"translate-server/services"
)

func IsInWhiteList(ip string, records []datamodels.IpTableRecord) bool {
	isInWhiteFlag := false
	for _, v := range records {
		if v.Type == "white" {
			arr := strings.Split(v.Ip, "-")
			if len(arr) == 2 {
				if ip >= arr[0] && ip <= arr[1] {
					isInWhiteFlag = true
					break
				}
			} else {
				arr = strings.Split(v.Ip, ",")
				for _, item := range arr {
					if item == ip {
						isInWhiteFlag = true
						break
					}
				}
			}
		}
	}
	return isInWhiteFlag
}

func IsInBlackList(ip string, records []datamodels.IpTableRecord) bool {
	for _, v := range records {
		if v.Type == "black" {
			arr := strings.Split(v.Ip, "-")
			if len(arr) == 2 {
				if ip >= arr[0] && ip <= arr[1] {
					return true
				}
			} else {
				arr = strings.Split(v.Ip, ",")
				for _, item := range arr {
					if item == ip {
						return true
					}
				}
			}
		}
	}
	return false
}

func IpAccessMiddleware(ctx iris.Context) {
	ipAddr := exnet.ClientIP(ctx.Request())
	service := services.NewIpTableService()
	tableType, err := service.GetIpTableType()
	if err != nil {
		log.Println(err)
		ctx.Next()
		return
	}
	if tableType == "" {
		ctx.Next()
		return
	}
	records, err := service.QueryRecords()
	if err != nil {
		log.Println(err)
		ctx.Next()
		return
	}
	if tableType == "black" {
		if IsInBlackList(ipAddr, records) {
			ctx.JSON(
				map[string]interface{}{
					"code": datamodels.HttpForbiddenIp,
					"msg":  datamodels.HttpForbiddenIp.String(),
				})
			return
		}
	} else {
		if IsInWhiteList(ipAddr, records) {
			ctx.Next()
			return
		} else {
			ctx.JSON(
				map[string]interface{}{
					"code": datamodels.HttpForbiddenIp,
					"msg":  datamodels.HttpForbiddenIp.String(),
				})
			return
		}
	}
	ctx.Next()
}
