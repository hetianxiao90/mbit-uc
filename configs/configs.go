package configs

import (
	"bytes"
	"fmt"
	"github.com/spf13/viper"
	"os"
	"uc/pkg/nacos"
)

const (
	ENV_NACOS_ENDPOINTS = "ENV_NACOS_ENDPOINTS"
	ENV_APP             = "ENV_APP"
)

var Config = new(MyConfig)

type App struct {
	Port int `yaml:"port"`
}
type Log struct {
	ErrorPath string `yaml:"error_path" mapstructure:"error_path"`
	InfoPath  string `yaml:"info_path" mapstructure:"info_path"`
	MaxAge    int    `yaml:"max_age" mapstructure:"max_age"`
	Rotation  int    `yaml:"rotation" mapstructure:"rotation"`
}
type DB struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DB       string `yaml:"db"`
}

type Mysql struct {
	Master *DB   `yaml:"master"`
	Slaves []*DB `yaml:"slaves"`
	Base   struct {
		Data            int
		MaxOpenConn     int `yaml:"max_open_conn" mapstructure:"max_open_conn"`
		MaxIdleConn     int `yaml:"max_idle_conn" mapstructure:"max_idle_conn"`
		ConnMaxLifeTime int `yaml:"conn_max_life_time" mapstructure:"conn_max_life_time"`
	}
}

type Email struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type Redis struct {
	Addr         []string `yaml:"addr"`
	Pass         string   `yaml:"pass"`
	Db           int      `yaml:"db"`
	MaxRetries   int      `yaml:"max_retries" mapstructure:"max_retries"`
	PoolSize     int      `yaml:"pool_size" mapstructure:"pool_size"`
	MinIdleConns int      `yaml:"min_idle_conns" mapstructure:"min_idle_conns"`
}

type RabbitMq struct {
	Host        string              `yaml:"host"`
	Port        int                 `yaml:"port"`
	Username    string              `yaml:"username"`
	Password    string              `yaml:"password"`
	MaxOpen     int                 `yaml:"max_open" mapstructure:"max_open"`
	MaxIdle     int                 `yaml:"max_idle" mapstructure:"max_idle"`
	MaxAttempts int                 `yaml:"max_attempts" mapstructure:"max_attempts"`
	Exchanges   *RabbitMqExchange   `yaml:"exchanges"`
	Queues      *RabbitMqQueues     `yaml:"queues"`
	RoutingKey  *RabbitMqRoutingKey `yaml:"routing_key" mapstructure:"routing_key"`
}
type RabbitMqExchange struct {
	User string `yaml:"user"`
}
type RabbitMqQueues struct {
	SendEmail string `yaml:"send_email" mapstructure:"send_email"`
}
type RabbitMqRoutingKey struct {
	Public string `yaml:"public"`
}

type Jwt struct {
	AccessTokenExpiredTime  int64  `json:"access_token_timeout" mapstructure:"access_token_expired_time"`
	RefreshTokenExpiredTime int64  `json:"refresh_token_timeout" mapstructure:"refresh_token_expired_time"`
	Secret                  string `json:"secret"`
}

type MyConfig struct {
	*App
	*Mysql
	*Email
	*Redis
	*RabbitMq
	*Log
	*Jwt
}

func Init() {

	//// 加载配置
	//viper.SetConfigFile("./configs/configs.yaml")
	//
	//// 监听配置
	//viper.WatchConfig()
	//
	//// 监听是否更改配置文件
	//viper.OnConfigChange(func(e fsnotify.Event) {
	//	if err := viper.Unmarshal(&Config); err != nil {
	//		panic(err)
	//	}
	//})
	//
	//if err := viper.ReadInConfig(); err != nil {
	//	panic(fmt.Errorf("ReadInConfig failed, err: %v", err))
	//}
	//if err := viper.Unmarshal(&Config); err != nil {
	//	panic(fmt.Errorf("unmarshal failed, err: %v", err))
	//}

	// 获取nacos节点
	endpoints, exist := os.LookupEnv(ENV_NACOS_ENDPOINTS)
	if !exist {
		panic("ENV_NACOS_ENDPOINTS not exsit")
	}

	// 获取当前环境
	envApp, exist := os.LookupEnv(ENV_APP)
	if !exist {
		panic("ENV_APP not exsit")
	}

	// 初始化Nacos配置
	nacosConfig := nacos.NewNacosConfig(endpoints, envApp, "user_config.yaml", "USER_GROUP")

	// 获取配置信息
	content, err := nacosConfig.GetConfig()
	if err != nil {
		panic(fmt.Errorf("GetConfig failed, err: %v", err))
	}

	viper.SetConfigType("yaml")
	if err := viper.ReadConfig(bytes.NewBuffer([]byte(content))); err != nil {
		panic(fmt.Errorf("ReadConfig failed, err: %v", err))
	}

	if err := viper.Unmarshal(&Config); err != nil {
		panic(fmt.Errorf("unmarshal failed, err: %v", err))
	}
}
