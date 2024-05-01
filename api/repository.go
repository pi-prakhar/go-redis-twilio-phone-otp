package api

import (
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/pi-prakhar/go-redis-twilio-phone-otp/internal/database"
	"github.com/pi-prakhar/go-redis-twilio-phone-otp/utils"
)

// store otp in cache
func storeInCache(key string, value any, expiry time.Duration) error {
	rdb := database.Client(0)
	ctx := database.Ctx

	err := rdb.Set(ctx, key, value, expiry).Err()
	if err != nil {
		utils.Log.Debug("Failed to store data in cache")
		return err
	}
	utils.Log.Info("Successfully Stored data in cache")
	return nil
}

// get data from cache
func getCachedData(key string) (any, error) {
	rdb := database.Client(0)
	ctx := database.Ctx
	cachedData, err := rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		utils.Log.Debug("Key not present in cache")
		return nil, nil
	} else if err != nil {
		utils.Log.Debug("Failed to get value from cache")
		return nil, err
	} else {
		utils.Log.Info("Key is present in cache")
		return cachedData, nil
	}
}

// decrement value of a key from cache
func decrementValueInCache(key string) error {
	rdb := database.Client(0)
	ctx := database.Ctx
	err := rdb.Decr(ctx, key).Err()
	if err != nil {
		utils.Log.Debug("Failed to decrement value from cache")
		return err
	} else {
		utils.Log.Info("Successfully decremented value from cache")
		return nil
	}
}

// increment value of a key from cache
func incrementValueInCache(key string) error {
	rdb := database.Client(0)
	ctx := database.Ctx
	err := rdb.Incr(ctx, key).Err()
	if err != nil {
		utils.Log.Debug("Failed to decrement value from cache")
		return err
	} else {
		utils.Log.Info("Successfully decremented value from cache")
		return nil
	}
}
