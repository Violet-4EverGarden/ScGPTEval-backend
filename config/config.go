package config

import (
	"github.com/fsnotify/fsnotify"
	"go.uber.org/zap"

	"github.com/spf13/viper"
)

// Conf 全局变量，用于保存程序的所有配置信息
var Conf = new(AppConfig)

// AppConfig 使用结构体变量保存配置信息；相比于直接用viper来保存，这种方式对程序员更友好
type AppConfig struct {
	Name         string `mapstructure:"name"`
	Mode         string `mapstructure:"mode"`
	Version      string `mapstructure:"version"`
	StartTime    string `mapstructure:"start_time"`
	MachineID    int64  `mapstructure:"machine_id"`
	Port         int    `mapstructure:"port"`
	*LogConfig   `mapstructure:"log"`
	*MySQLConfig `mapstructure:"mysql"`
	*RedisConfig `mapstructure:"redis"`
}

type LogConfig struct {
	Level      string `mapstructure:"level"`
	Filename   string `mapstructure:"filename"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxAge     int    `mapstructure:"max_age"`
	MaxBackups int    `mapstructure:"max_backups"`
}

type MySQLConfig struct {
	Host         string `mapstructure:"host"`
	User         string `mapstructure:"user"`
	Password     string `mapstructure:"password"`
	DbName       string `mapstructure:"db_name"`
	DbParams     string `mapstructure:"db_params"`
	Port         int    `mapstructure:"port"`
	MaxOpenConns int    `mapstructure:"max_open_conns"`
	MaxIdleConns int    `mapstructure:"max_idle_conns"`
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Password string `mapstructure:"password"`
	Port     int    `mapstructure:"port"`
	DB       int    `mapstructure:"db"`
	PoolSize int    `mapstructure:"pool_size"`
}

func Init() (err error) {
	viper.SetConfigName("config_local") // 指定配置文件名称(无扩展名)
	viper.AddConfigPath(".")            // 查找配置文件所在的路径(使用相对路径)
	viper.AddConfigPath("./conf")
	// 查找并读取配置文件
	if err = viper.ReadInConfig(); err != nil {
		zap.L().Error("配置文件读取失败：", zap.Error(err))
		return
	}
	// 将读取到的配置信息反序列化到 Conf 变量中
	if err = viper.Unmarshal(Conf); err != nil {
		zap.L().Error("配置信息反序列化失败：", zap.Error(err))
	}
	// 实时监控配置文件变化并更新配置信息
	viper.WatchConfig()
	viper.OnConfigChange(func(in fsnotify.Event) {
		zap.L().Warn("配置文件修改了...")
		// fmt.Println("配置文件修改了...")
		if err = viper.Unmarshal(Conf); err != nil {
			zap.L().Error("配置信息反序列化失败：", zap.Error(err))
		}
	})
	return
}
