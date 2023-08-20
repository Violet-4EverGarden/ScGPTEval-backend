package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"scgptEval/config"
	"scgptEval/controllers"
	"scgptEval/dao/mysql"
	"scgptEval/dao/redis"
	"scgptEval/logger"
	"scgptEval/pkg/snowflake"
	"scgptEval/routes"
	"syscall"
	"time"

	"github.com/spf13/viper"

	"go.uber.org/zap"
)

func main() {
	// 1. 加载配置文件
	if err := config.Init(); err != nil {
		fmt.Printf("init config failed, err:%v\n", err)
		return
	}
	// 2. 初始化日志
	if err := logger.Init(config.Conf.LogConfig, config.Conf.Mode); err != nil {
		fmt.Printf("init logger failed, err:%v\n", err)
		return
	}
	defer zap.L().Sync()
	zap.L().Debug("logger init success...") // 通过zap.L调用全局Logger实例
	// 3. 初始化MySQL
	if err := mysql.Init(config.Conf.MySQLConfig); err != nil {
		fmt.Printf("init mysql failed, err:%v\n", err)
		return
	}
	defer mysql.Close() // 延迟关闭mysql连接
	// 4. 初始化Redis连接
	if err := redis.Init(config.Conf.RedisConfig); err != nil {
		fmt.Printf("init redis failed, err:%v\n", err)
		return
	}
	defer redis.Close() // 延迟关闭redis连接
	// 雪花算法生成用户id和帖子id
	if err := snowflake.Init(config.Conf.StartTime, config.Conf.MachineID); err != nil {
		fmt.Printf("init snowflake failed, err:%v\n", err)
		return
	}
	// 初始化gin框架内置的参数校验器使用的翻译器
	if err := controllers.InitTrans("zh"); err != nil {
		fmt.Printf("init validator trans failed, err:%v\n", err)
		return
	}
	// 5. 注册路由
	r := routes.SetUp()
	// 6. 启动服务（优雅关机、平滑重启）
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", viper.GetInt("port")),
		Handler: r,
	}
	go func() {
		// 开启一个goroutine启动服务
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()
	// 等待中断信号来优雅地关闭服务器，为关闭服务器操作设置一个5秒的超时
	quit := make(chan os.Signal, 1) // 创建一个接收信号的通道
	// kill 默认会发送 syscall.SIGTERM 信号
	// kill -2 发送 syscall.SIGINT 信号，我们常用的Ctrl+C就是触发系统SIGINT信号
	// kill -9 发送 syscall.SIGKILL 信号，但是不能被捕获，所以不需要添加它
	// signal.Notify把收到的 syscall.SIGINT或syscall.SIGTERM 信号转发给quit
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM) // 此处不会阻塞
	<-quit                                               // 阻塞在此，当接收到上述两种信号时才会往下执行
	zap.L().Info("Shutdown Server ...")
	// 创建一个5秒超时的context
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// 5秒内优雅关闭服务（将未处理完的请求处理完再关闭服务），超过5秒就超时退出
	if err := srv.Shutdown(ctx); err != nil {
		zap.L().Fatal("Server Shutdown: ", zap.Error(err))
	}

	zap.L().Info("Server exiting")
}
