package mysql

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
	"scgptEval/config"
	"scgptEval/models"

	_ "github.com/jinzhu/gorm/dialects/mysql"
)

var db *gorm.DB

// 初始化mysql服务
func Init(cfg *config.MySQLConfig) (err error) {
	// DSN:Data Source Name, 通过viper读取配置文件中的mysql信息
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?%s",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		"",
		cfg.DbParams,
	)
	db, err = gorm.Open("mysql", dsn)
	if err != nil {
		zap.L().Error("connect DB failed", zap.Error(err))
		return
	}
	// 检查指定dbName的数据库是否存在，若不存在则自动创建
	checkDBExist(cfg.DbName)
	db.Close()

	// 使用刚刚创建的数据库名称构建新的数据库连接url，并重新实例化数据库引擎
	dsn = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?%s",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.DbName,
		cfg.DbParams,
	)
	db, err = gorm.Open("mysql", dsn)
	if err != nil {
		zap.L().Error("connect DB failed", zap.Error(err))
		return
	}
	// db.LogMode(true)
	db.DB().SetMaxOpenConns(cfg.MaxOpenConns) // 设置与数据库建立连接的最大数目
	db.DB().SetMaxIdleConns(cfg.MaxIdleConns) // 设置连接池中的最大闲置连接数

	// ping一下查看mysql是否已连接上
	if err = db.DB().Ping(); err != nil {
		panic(err)
	}
	// 自动同步数据库表
	syncTable()
	return
}

func checkDBExist(DbName string) {
	// 没有数据库就自动新建
	res := db.Exec("CREATE DATABASE IF NOT EXISTS " + DbName + " DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci")
	if res.Error != nil {
		fmt.Println("数据库已存在 或 新建数据库失败...", res.Error.Error())
		return
	}
}

// 自动同步数据库，没有表就自动新建，必要时打开
func syncTable() {
	res := db.AutoMigrate(
		&models.User{},
		&models.Quiz{},
		&models.Record{},
	)
	if res.Error != nil {
		fmt.Println("同步数据库字段失败", res.Error.Error())
		return
	}
}

// 由于全局变量db是不对外暴露的，因此需要封装一个Close函数以便在main.go中调用db.Close()
func Close() {
	_ = db.Close()
}
