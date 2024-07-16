package mysql

import (
	"database/sql"
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"
	"log"
	"time"
	"uc/pkg/nacos"
)

var DBG = new(gorm.DB)
var DBS = new(sql.DB)

// DBConfig 配置数据库连接信息
type DBConfig struct {
	Username string
	Password string
	Host     string
	Port     int
	Database string
}

func Init() {
	dsn := getDSN(&DBConfig{
		Username: nacos.Config.Mysql.Master.User,
		Password: nacos.Config.Mysql.Master.Password,
		Host:     nacos.Config.Mysql.Master.Host,
		Port:     nacos.Config.Mysql.Master.Port,
		Database: nacos.Config.Mysql.Master.DB,
	})

	db, err := gorm.Open(mysql.New(mysql.Config{
		DSN: dsn,
	}))

	if err != nil {
		log.Fatal("Could not connect to the master database:", err)
	}

	var replicas []gorm.Dialector
	for _, s := range nacos.Config.Slaves {
		dsnSlaves := getDSN(&DBConfig{
			Username: s.User,
			Password: s.Password,
			Host:     s.Host,
			Port:     s.Port,
			Database: s.DB,
		})
		replicas = append(replicas, mysql.New(mysql.Config{DSN: dsnSlaves}))
	}
	err = db.Use(
		dbresolver.Register(dbresolver.Config{
			Sources: []gorm.Dialector{mysql.New(mysql.Config{
				DSN: dsn,
			})},
			Replicas: replicas,
			Policy:   dbresolver.RandomPolicy{},
		}).
			SetMaxOpenConns(nacos.Config.Mysql.Base.MaxOpenConn).
			SetMaxIdleConns(nacos.Config.Mysql.Base.MaxIdleConn).
			SetConnMaxLifetime(time.Duration(nacos.Config.Mysql.Base.ConnMaxLifeTime)),
	)
	if err != nil {
		log.Fatal("Could not connect to the replicas database:", err)
	}
	DBG = db
	DBS, err = db.DB()
}

// getDSN 生成DSN（数据源名称）
func getDSN(cfg *DBConfig) string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=%t&loc=%s",
		cfg.Username,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Database,
		true,
		"Local")
}
