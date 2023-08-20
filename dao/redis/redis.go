package redis

import (
	"fmt"
	"scgptEval/config"

	"github.com/go-redis/redis"
)

// 声明一个全局的rdb变量
var rdb *redis.Client

// 初始化连接
func Init(cfg *config.RedisConfig) (err error) {
	rdb = redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%d",
			cfg.Host,
			cfg.Port,
		),
		Password: cfg.Password, // no password set
		DB:       cfg.DB,       // use default DB
		PoolSize: cfg.PoolSize, // 最大连接数
	})

	_, err = rdb.Ping().Result()
	return
}

// rdb变量不对外暴露，因此需要封装一个Close()方法以便main.go在执行结束后调用rdb.Close()及时关闭redis连接
func Close() {
	_ = rdb.Close()
}
