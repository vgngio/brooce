// +build !cluster

package redis

import (
	"context"
	"log"
	"sync"
	"time"

	"brooce/config"

	"github.com/go-redis/redis/v8"
)

var ctx = context.Background()

var redisClient *redis.Client
var once sync.Once

func Get() (*redis.Client, context.Context) {
	once.Do(func() {
		threads := len(config.Threads) + 10

		var redisOptions *redis.Options

		if len(config.Config.Redis.URL) > 0 {
			var err error

			redisOptions, err = redis.ParseURL(config.Config.Redis.URL)

			if err != nil {
				log.Println("Invalid Redis URL provided.", err)

				return
			}
		} else {
			redisOptions = &redis.Options{
				Addr:     config.Config.Redis.Host,
				Password: config.Config.Redis.Password,
				DB:       config.Config.Redis.DB,
			}
		}

		redisOptions.MaxRetries = 10
		redisOptions.PoolSize = threads
		redisOptions.DialTimeout = 5 * time.Second
		redisOptions.ReadTimeout = 30 * time.Second
		redisOptions.WriteTimeout = 5 * time.Second
		redisOptions.PoolTimeout = 1 * time.Second

		redisClient = redis.NewClient(redisOptions)

		for {
			err := redisClient.Ping(ctx).Err()
			if err == nil {
				break
			}

			host := config.Config.Redis.Host

			if len(config.Config.Redis.URL) > 0 {
				host = config.Config.Redis.URL
			}

			log.Println("Can't reach redis at", host, "-- are your redis addr and password right?", err)
			time.Sleep(5 * time.Second)
		}
	})

	return redisClient, ctx
}

func FlushList(src, dst string) (err error) {
	redisClient, ctx := Get()
	for err == nil {
		_, err = redisClient.RPopLPush(ctx, src, dst).Result()
	}

	if err == redis.Nil {
		err = nil
	}

	return
}

func ScanKeys(match string) (keys []string, err error) {
	redisClient, ctx := Get()

	iter := redisClient.Scan(ctx, 0, match, 10000).Iterator()
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}
	err = iter.Err()

	return
}
