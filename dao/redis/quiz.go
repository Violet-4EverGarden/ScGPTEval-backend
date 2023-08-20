package redis

import (
	"errors"
	"fmt"
	"github.com/go-redis/redis"
	"strconv"
	"time"
)

// GetUserAmount 获取指定id用户的刷题数量
func GetUserAmount(userID string) (int64, error) {
	score, err := rdb.ZScore(KeyUserScoreZSet, userID).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			// member不存在说明用户未刷过题
			return 0, nil
		}
		// 其它错误
		return 0, err
	}
	return int64(score), nil
}

// GetRanking 获取刷题量排行榜：前size个，降序
func GetRanking(size int64) (rank []map[string]string, err error) {
	users, err := rdb.ZRevRangeWithScores(KeyUserScoreZSet, 0, size-1).Result()
	if err != nil {
		return nil, err
	}

	rank = make([]map[string]string, 0, len(users))
	pipeline := rdb.Pipeline()
	for _, user := range users {
		pipeline.HGetAll(KeyUserInfoHashPrefix + fmt.Sprint(user.Member))
	}
	cmds, err := pipeline.Exec()
	if err != nil {
		return nil, err
	}
	for i, cmd := range cmds {
		r := cmd.(*redis.StringStringMapCmd).Val()
		r["id"] = users[i].Member.(string)
		r["score"] = strconv.FormatFloat(users[i].Score, 'f', 0, 64)
		// 时间戳 --> 正常时间

		rank = append(rank, r)
	}
	return rank, nil
}

// SubmitQuiz 提交题目
func SubmitQuiz(userID, quizID string) (err error) {
	// 更新存储用户的刷题总量zset、用户刷过的题set、用户的刷题信息hash（更新最后刷题时间戳）
	now := time.Now().Unix()
	// redis事务操作
	pipeline := rdb.TxPipeline()
	pipeline.ZIncrBy(KeyUserScoreZSet, 1, userID) // 用户刷题量增1
	pipeline.SAdd(KeyUserQuizSetPrefix+userID, quizID)
	pipeline.HSet(KeyUserInfoHashPrefix+userID, "final_qa_time", now)

	if _, err = pipeline.Exec(); err != nil {
		return err
	}
	return nil
}

// GetUntriedQuiz 获取用户未刷过的题目ids
func GetUntriedQuiz(userID string) (quizIds []string, err error) {
	// 执行SDIFF进行差集运算，获得用户为刷过的题目id集合
	res, err := rdb.SDiff(KeyQuizBaseSet, KeyUserQuizSetPrefix+userID).Result()
	if err != nil {
		return nil, err
	}
	return res, nil
}
