package main

import (
	"fmt"

	"gopkg.in/redis.v5"
)

var RedisClient *redis.Client

func InitRedis() {
	// TODO: make password and db configurable
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", config.Redis.Host, config.Redis.Port),
		Password: "",
		DB:       0,
	})

	_, err := RedisClient.Ping().Result()
	if err != nil {
		panic(err)
	}
}

func DeInitRedis() {
	RedisClient.Close()
}
