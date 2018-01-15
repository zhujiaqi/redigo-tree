package redigotree

import (
	"github.com/garyburd/redigo/redis"
	log "github.com/sirupsen/logrus"
	"time"
)

var (
	redisClient   *redis.Pool
	redis_host    string
	redis_db      int
	redis_network string
	password      string
	maxIdle       int
	maxActive     int
	idleTimeout   time.Duration
)

func init() {
	redis_host = "127.0.0.1:6379"
	redis_db = 0
	redis_network = "tcp"
	password = ""
	maxIdle = 1
	maxActive = 10
	idleTimeout = 10
	redisClient = &redis.Pool{
		MaxIdle:     maxIdle,
		MaxActive:   maxActive,
		IdleTimeout: idleTimeout * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial(redis_network, redis_host)
			if err != nil {
				log.Fatal("Redis Pool Failed To Create")
			}
			if password != "" {
				if _, err := c.Do("AUTH", password); err != nil {
					c.Close()
					return nil, err
				}
			}
			c.Do("SELECT", redis_db)
			return c, nil
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
		Wait: true,
	}
}
