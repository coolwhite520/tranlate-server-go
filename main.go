package main

import (
	log "github.com/sirupsen/logrus"
	"time"
	"translate-server/http"
	"translate-server/logcuthook"
	"translate-server/loghook"
)


func init() {
	logcuthook.ConfigLocalFilesystemLogger("./logs", "mylog", time.Hour*24*60, time.Hour*24)
	log.AddHook(loghook.NewContextHook())
	log.SetFormatter(&log.TextFormatter{TimestampFormat: "2006-01-02 15:04:05"})
}

func main()  {
	log.WithFields(log.Fields{
		"ServerResponse": true,
		"ReqUrl":         "baidu.com",
	}).Info("gou dong xi ")
	http.StartIntServer()
}