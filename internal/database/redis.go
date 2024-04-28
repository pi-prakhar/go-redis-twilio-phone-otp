package database

import (
	"context"

	"github.com/go-redis/redis/v8"
	"github.com/pi-prakhar/go-redis-twilio-phone-otp/utils"
	"github.com/pi-prakhar/utils/loader"
)

var Ctx = context.Background()

func Client(dbNo int) *redis.Client {
	address, err := loader.GetValueFromConf("redis-db-address")
	if err != nil {
		utils.Log.Error("Failed to get value from conf", err)
	}

	password, err := loader.GetValueFromEnv("REDIS_DB_PASSWORD")
	if err != nil {
		utils.Log.Error("Failed to get value from env", err)
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     address,
		Password: password,
		DB:       dbNo,
	})

	return rdb
}
