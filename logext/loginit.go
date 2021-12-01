package logext

import (
	log "github.com/sirupsen/logrus"
	"time"
)

func init() {
	ConfigLocalFilesystemLogger("./logs", "logext", time.Hour*24*60, time.Hour*24)
	log.AddHook(NewContextHook())
	log.SetFormatter(&log.TextFormatter{TimestampFormat: "2006-01-02 15:04:05"})
}

