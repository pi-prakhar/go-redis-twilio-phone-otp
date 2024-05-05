package api

import (
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/pi-prakhar/go-redis-twilio-phone-otp/internal/database"
	"github.com/pi-prakhar/go-redis-twilio-phone-otp/utils"
)

func storeInCache(key string, value any, expiry time.Duration) error {
	rdb := database.Client(0)
	ctx := database.Ctx

	err := rdb.Set(ctx, key, value, expiry).Err()
	if err != nil {
		utils.Log.Debug("Error : Failed to store data in cache")
		return err
	}
	utils.Log.Debug("Successfully Stored data in cache")
	return nil
}

func storeInCacheNoExpiry(key string, value any) error {
	rdb := database.Client(0)
	ctx := database.Ctx

	err := rdb.Set(ctx, key, value, 0).Err()
	if err != nil {
		utils.Log.Debug("Error : Failed to store data with no expiry in cache")
		return err
	}
	utils.Log.Debug("Successfully Stored data with no expiry in cache")
	return nil
}

func getCachedData(key string) (any, error) {
	rdb := database.Client(0)
	ctx := database.Ctx
	cachedData, err := rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		utils.Log.Debug("Data not present in cache")
		return nil, nil
	} else if err != nil {
		utils.Log.Debug("Error : Failed to fetch data from cache")
		return nil, err
	} else {
		utils.Log.Debug("Successfully fetched data from cache")
		return cachedData, nil
	}
}

func getTTLData(key string) (time.Duration, error) {
	rdb := database.Client(0)
	ctx := database.Ctx
	ttl, err := rdb.TTL(ctx, key).Result()
	if err != nil {
		utils.Log.Debug("Error : Failed to fetch ttl from cache")
		return ttl, err
	}
	if ttl.Seconds() == -1 {
		utils.Log.Debug("Key does not expire")
		return ttl, nil
	} else if ttl.Seconds() == -2 {
		utils.Log.Debug("Error : Key does not exists")
		return ttl, nil
	}
	utils.Log.Debug("Successfully fetched ttl from cache")
	return ttl, nil
}

func deleteDataFromCache(key string) error {
	rdb := database.Client(0)
	ctx := database.Ctx
	err := rdb.Del(ctx, key).Err()
	if err == redis.Nil {
		utils.Log.Debug("key not present in cache")
		return nil
	}
	if err != nil {
		utils.Log.Debug("Error : Failed deleting data from cache")
		return err
	}
	utils.Log.Debug("Successfully deleted data from cache")
	return nil
}

func decrementValueInCache(key string) error {
	rdb := database.Client(0)
	ctx := database.Ctx
	err := rdb.Decr(ctx, key).Err()
	if err != nil {
		utils.Log.Debug("Error : Failed to decrement data in cache")
		return err
	} else {
		utils.Log.Info("Successfully decremented data in cache")
		return nil
	}
}
