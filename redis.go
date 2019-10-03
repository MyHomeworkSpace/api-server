package main

import (
	"fmt"

	"github.com/MyHomeworkSpace/api-server/config"

	"gopkg.in/redis.v5"
)

// RedisClient is a pointer to the current connection to Redis
var RedisClient *redis.Client

func initRedis() {
	// TODO: make password and db configurable
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", config.GetCurrent().Redis.Host, config.GetCurrent().Redis.Port),
		Password: "",
		DB:       0,
	})

	_, err := RedisClient.Ping().Result()
	if err != nil {
		panic(err)
	}
}

func deinitRedis() {
	RedisClient.Close()
}
