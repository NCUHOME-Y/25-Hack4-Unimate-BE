package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/NCUHOME-Y/25-Hack4-Unimate-BE/internal/app/model"
	"github.com/NCUHOME-Y/25-Hack4-Unimate-BE/internal/app/repository"
)

// Redis 排行榜服务 - 使用缓存加速排行榜查询

const (
	// Redis key 定义
	RankingKeyCounts       = "ranking:counts"      // 积分排行榜
	RankingKeyMonthLearn   = "ranking:month_learn" // 月学习时间排行榜
	RankingKeyDaka         = "ranking:daka"        // 打卡数排行榜
	RankingKeyFlagNumber   = "ranking:flag_number" // Flag数量排行榜
	RankingCacheTTL        = 5 * time.Minute       // 缓存过期时间：5分钟
	RankingCacheTTLSeconds = int64(5 * 60)         // 缓存过期时间（秒）
)

// ==================== 缓存更新函数 ====================

// 更新所有排行榜缓存
func RefreshAllRankings() error {
	if err := RefreshRankingCounts(); err != nil {
		log.Printf("[error] 更新积分排行榜失败: %v", err)
	}
	if err := RefreshRankingMonthLearn(); err != nil {
		log.Printf("[error] 更新月学习时间排行榜失败: %v", err)
	}
	if err := RefreshRankingDaka(); err != nil {
		log.Printf("[error] 更新打卡数排行榜失败: %v", err)
	}
	if err := RefreshRankingFlagNumber(); err != nil {
		log.Printf("[error] 更新Flag数量排行榜失败: %v", err)
	}
	return nil
}

// 更新积分排行榜缓存
func RefreshRankingCounts() error {
	// ① 从 MySQL 读取数据
	users, err := repository.GetUserByCount()
	if err != nil {
		return err
	}

	// ② 将数据序列化为 JSON
	data, err := json.Marshal(users)
	if err != nil {
		return err
	}

	// ③ 存入 Redis（带过期时间）
	err = repository.RedisSetEx(RankingKeyCounts, string(data), RankingCacheTTL)
	log.Printf("[info] 已更新积分排行榜缓存，共 %d 条记录", len(users))
	return err
}

// 更新月学习时间排行榜缓存
func RefreshRankingMonthLearn() error {
	users, err := repository.GetUserByMonthLearnTime()
	if err != nil {
		return err
	}

	data, err := json.Marshal(users)
	if err != nil {
		return err
	}

	err = repository.RedisSetEx(RankingKeyMonthLearn, string(data), RankingCacheTTL)
	log.Printf("[info] 已更新月学习时间排行榜缓存，共 %d 条记录", len(users))
	return err
}

// 更新打卡数排行榜缓存
func RefreshRankingDaka() error {
	users, err := repository.GetUserByDaka()
	if err != nil {
		return err
	}

	data, err := json.Marshal(users)
	if err != nil {
		return err
	}

	err = repository.RedisSetEx(RankingKeyDaka, string(data), RankingCacheTTL)
	log.Printf("[info] 已更新打卡数排行榜缓存，共 %d 条记录", len(users))
	return err
}

// 更新Flag数量排行榜缓存
func RefreshRankingFlagNumber() error {
	users, err := repository.GetUserByFlagNumber()
	if err != nil {
		return err
	}

	data, err := json.Marshal(users)
	if err != nil {
		return err
	}

	err = repository.RedisSetEx(RankingKeyFlagNumber, string(data), RankingCacheTTL)
	log.Printf("[info] 已更新Flag数量排行榜缓存，共 %d 条记录", len(users))
	return err
}

// ==================== 查询函数（优先读取缓存） ====================

// 获取积分排行榜（先查 Redis，没有则查 MySQL）
func GetRankingCountsWithCache() ([]model.User, error) {
	return getRankingWithCache(
		RankingKeyCounts,
		func() ([]model.User, error) { return repository.GetUserByCount() },
		RefreshRankingCounts,
	)
}

// 获取月学习时间排行榜
func GetRankingMonthLearnWithCache() ([]model.User, error) {
	return getRankingWithCache(
		RankingKeyMonthLearn,
		func() ([]model.User, error) { return repository.GetUserByMonthLearnTime() },
		RefreshRankingMonthLearn,
	)
}

// 获取打卡数排行榜
func GetRankingDakaWithCache() ([]model.User, error) {
	return getRankingWithCache(
		RankingKeyDaka,
		func() ([]model.User, error) { return repository.GetUserByDaka() },
		RefreshRankingDaka,
	)
}

// 获取Flag数量排行榜
func GetRankingFlagNumberWithCache() ([]model.User, error) {
	return getRankingWithCache(
		RankingKeyFlagNumber,
		func() ([]model.User, error) { return repository.GetUserByFlagNumber() },
		RefreshRankingFlagNumber,
	)
}

// 通用缓存查询函数
func getRankingWithCache(
	redisKey string,
	mysqlQuery func() ([]model.User, error),
	refreshFunc func() error,
) ([]model.User, error) {
	// ① 尝试从 Redis 读取
	cachedData, err := repository.RedisGet(redisKey)
	if err == nil && cachedData != "" {
		var users []model.User
		if err := json.Unmarshal([]byte(cachedData), &users); err == nil {
			log.Printf("[cache hit] 从 Redis 读取 %s", redisKey)
			return users, nil
		}
	}

	// ② Redis 中没有或过期，从 MySQL 读取
	log.Printf("[cache miss] %s 不在缓存中，从 MySQL 读取", redisKey)
	users, err := mysqlQuery()
	if err != nil {
		return nil, err
	}

	// ③ 后台异步更新缓存，不影响当前请求
	go func() {
		if err := refreshFunc(); err != nil {
			log.Printf("[error] 更新缓存失败: %v", err)
		}
	}()

	return users, nil
}

// 使用 Redis 有序集合实现排行榜
// 优点：排序、分页、分数查询都很快

// 更新用户积分到 Redis 有序集合


// 从 Redis 有序集合获取排行榜（支持分页）
func GetRankingFromRedisZSet(key string, start, stop int64) ([]map[string]interface{}, error) {
	// 从高分到低分获取
	results, err := repository.RedisClient.ZRevRangeWithScores(context.Background(), key, start, stop).Result()
	if err != nil {
		return nil, err
	}

	ranking := make([]map[string]interface{}, len(results))
	for i, result := range results {
		ranking[i] = map[string]interface{}{
			"rank":    start + int64(i) + 1,
			"user_id": result.Member,
			"score":   result.Score,
		}
	}
	return ranking, nil
}

// 获取用户在排行榜中的排名
func GetUserRankInRedis(key string, userID uint) (int64, error) {
	// 返回用户排名（从 0 开始，需要 +1）
	rank, err := repository.RedisClient.ZRevRank(context.Background(), key, fmt.Sprintf("%d", userID)).Result()
	if err != nil {
		return 0, err
	}
	return rank + 1, nil
}

// 定时刷新所有排行榜缓存（可在启动时或定时任务中调用）
func StartRankingCacheRefresh() {
	// 初始刷新
	RefreshAllRankings()

	// 定时刷新（每 5 分钟）
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	go func() {
		for range ticker.C {
			log.Println("[info] 定时刷新排行榜缓存...")
			if err := RefreshAllRankings(); err != nil {
				log.Printf("[error] 刷新排行榜缓存失败: %v", err)
			}
		}
	}()
}
