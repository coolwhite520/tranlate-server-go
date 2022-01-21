package datamodels

import (
	"fmt"
	"github.com/go-redis/redis"
	log "github.com/sirupsen/logrus"
	"time"
)

var ClientRedis *redis.Client

const (
	RedisNetwork  = "tcp"
	RedisHost     = "127.0.0.1"
	RedisPort     = "6379"
	RedisPassword = ""
	RedisDb       = 0
)

func init() {
	options := redis.Options{
		Network:            RedisNetwork,
		Addr:               fmt.Sprintf("%s:%s", RedisHost, RedisPort),
		Dialer:             nil,
		OnConnect:          nil,
		Password:           RedisPassword,
		DB:                 RedisDb,
		MaxRetries:         0,
		MinRetryBackoff:    0,
		MaxRetryBackoff:    0,
		DialTimeout:        0,
		ReadTimeout:        0,
		WriteTimeout:       0,
		PoolSize:           0,
		MinIdleConns:       0,
		MaxConnAge:         0,
		PoolTimeout:        0,
		IdleTimeout:        0,
		IdleCheckFrequency: 0,
		TLSConfig:          nil,
	}
	// 新建一个client
	ClientRedis = redis.NewClient(&options)
	// close
	// defer ClientRedis.Close()
	ping := ClientRedis.Ping()
	for i:=0; i < 100; i++ {
		time.Sleep(1 * time.Second)
		result, err := ping.Result()
		if err != nil {
			log.Errorln(err)
			continue
		}
		log.Info(result)
		break
	}
}

func SetRedisString(key string, val string, t time.Duration) {
	// 添加string
	ClientRedis.Set(key, val, t)
}

func GetRedisString(key string) string {
	cmd := ClientRedis.Get(key)
	if cmd != nil {
		return cmd.Val()
	}
	return ""
}
