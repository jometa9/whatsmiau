package services

import (
	"crypto/tls"
	"sync"

	"github.com/verbeux-ai/whatsmiau/env"
	"github.com/verbeux-ai/whatsmiau/interfaces"
	"github.com/verbeux-ai/whatsmiau/repositories/instances"
	"golang.org/x/net/context"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

var (
	redisInstance *redis.Client
	redisError    error
	sqliteRepo    interfaces.InstanceRepository
	sqliteOnce    sync.Once
)

func Redis() *redis.Client {
	if redisInstance == nil && redisError == nil {
		instance, err := NewRedis()
		if err != nil {
			zap.L().Warn("failed to start redis, will use SQLite repository", zap.Error(err))
			redisError = err
			return nil
		}

		redisInstance = instance
	}

	return redisInstance
}

func RedisAvailable() bool {
	return Redis() != nil
}

// getSQLiteRepository returns a SQLite repository instance (singleton)
func getSQLiteRepository() interfaces.InstanceRepository {
	sqliteOnce.Do(func() {
		// Extract the database path from DBURL
		// DBURL format: "file:data.db?_foreign_keys=on" or just "data.db"
		dbPath := env.Env.DBURL
		if dbPath == "" {
			dbPath = "file:data.db?_foreign_keys=on"
		}
		
		// Remove query parameters for SQLite connection
		// SQLite driver handles the full connection string
		repo, err := instances.NewSQLite(dbPath)
		if err != nil {
			zap.L().Error("failed to create SQLite repository, falling back to memory", zap.Error(err))
			sqliteRepo = instances.NewMemory()
		} else {
			sqliteRepo = repo
			if env.Env.DebugMode {
				zap.L().Info("using SQLite repository for instances", zap.String("dbPath", dbPath))
			}
		}
	})
	return sqliteRepo
}

// GetInstanceRepository returns the instance repository, using Redis if available, otherwise SQLite
func GetInstanceRepository() interfaces.InstanceRepository {
	if RedisAvailable() {
		return instances.NewRedis(Redis())
	}
	return getSQLiteRepository()
}

func NewRedis() (*redis.Client, error) {
	opt := &redis.Options{
		Addr:     env.Env.RedisURL,
		Password: env.Env.RedisPassword,
		DB:       0,
	}

	if env.Env.RedisTLS {
		opt.TLSConfig = &tls.Config{}
	}

	client := redis.NewClient(opt)
	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}

	return client, nil
}
