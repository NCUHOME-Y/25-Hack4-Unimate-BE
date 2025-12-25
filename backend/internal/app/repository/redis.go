package repository

import (
	"context"
	"os"
	"strconv"
	"time"

	utils "github.com/NCUHOME-Y/25-Hack4-Unimate-BE/util"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

var (
	RedisClient *redis.Client
	Ctx         = context.Background() // 导出Ctx供其他包使用
)

// 初始化 Redis 连接
func RedisConnect() {
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	redisPassword := os.Getenv("REDIS_PASSWORD")
	redisDB := 0
	if dbStr := os.Getenv("REDIS_DB"); dbStr != "" {
		if db, err := strconv.Atoi(dbStr); err == nil {
			redisDB = db
		}
	}

	RedisClient = redis.NewClient(&redis.Options{
		Addr:         redisAddr,
		Password:     redisPassword,
		DB:           redisDB,
		PoolSize:     10,              // 最大连接池大小
		MinIdleConns: 5,               // 最小空闲连接数
		MaxRetries:   3,               // 最大重试次数
		PoolTimeout:  4 * time.Second, // 连接池超时时间
		ReadTimeout:  3 * time.Second, // 读取超时时间
		WriteTimeout: 3 * time.Second, // 写入超时时间
	})

	// 测试连接
	if err := RedisClient.Ping(Ctx).Err(); err != nil {
		utils.LogError("Redis 连接失败", logrus.Fields{"error": err})
		return
	}

	utils.LogInfo("✅ Redis 连接成功", map[string]interface{}{
		"addr": redisAddr,
		"db":   redisDB,
	})
}

// 获取 Redis 值
func RedisGet(key string) (string, error) {
	val, err := RedisClient.Get(Ctx, key).Result()
	if err == redis.Nil {
		return "", nil
	}
	return val, err
}

// 设置 Redis 值（无过期时间）
func RedisSet(key string, value interface{}) error {
	return RedisClient.Set(Ctx, key, value, 0).Err()
}

// 设置 Redis 值（带过期时间）
func RedisSetEx(key string, value interface{}, expiration time.Duration) error {
	return RedisClient.Set(Ctx, key, value, expiration).Err()
}

// 删除 Redis 键
func RedisDel(keys ...string) error {
	return RedisClient.Del(Ctx, keys...).Err()
}

// 检查键是否存在
func RedisExists(keys ...string) (int64, error) {
	return RedisClient.Exists(Ctx, keys...).Result()
}

// 获取过期时间
func RedisTTL(key string) (time.Duration, error) {
	return RedisClient.TTL(Ctx, key).Result()
}

// 设置过期时间
func RedisExpire(key string, expiration time.Duration) error {
	return RedisClient.Expire(Ctx, key, expiration).Err()
}

// 获取哈希字段值
func RedisHGet(key string, field string) (string, error) {
	val, err := RedisClient.HGet(Ctx, key, field).Result()
	if err == redis.Nil {
		return "", nil
	}
	return val, err
}

// 设置哈希字段值
func RedisHSet(key string, field string, value interface{}) error {
	return RedisClient.HSet(Ctx, key, field, value).Err()
}

// 获取所有哈希字段
func RedisHGetAll(key string) (map[string]string, error) {
	return RedisClient.HGetAll(Ctx, key).Result()
}

// 删除哈希字段
func RedisHDel(key string, fields ...string) error {
	return RedisClient.HDel(Ctx, key, fields...).Err()
}

// 推送到列表
func RedisPush(key string, values ...interface{}) error {
	return RedisClient.RPush(Ctx, key, values...).Err()
}

// 获取列表范围
func RedisRange(key string, start int64, stop int64) ([]string, error) {
	return RedisClient.LRange(Ctx, key, start, stop).Result()
}

// 关闭 Redis 连接
func RedisClose() error {
	if RedisClient != nil {
		return RedisClient.Close()
	}
	return nil
}
