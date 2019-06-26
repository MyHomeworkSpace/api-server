package data

import (
	"database/sql"

	"gopkg.in/redis.v5"
)

var DB *sql.DB
var RedisClient *redis.Client
