package manager

import (
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// DSN ...
type DSN struct {
	User     string            // Username
	Password string            // Password (requires User)
	Net      string            // Network type
	Addr     string            // Network address (requires Net)
	DBName   string            // Database name
	Params   map[string]string // Connection parameters
}

type DSNParser interface {
	GetDialector(dsn string) gorm.Dialector
	ParseDSN(dsn string) (cfg *DSN, err error)
	Scheme() string
	// NamingStrategy gorm naming strategy
	// 该方法主要用于达梦数据库
	// 达梦数据库的表名和字段名都是大写的，gorm默认的策略是小写
	// 所以需要该方法来设置gorm的达梦命名策略
	NamingStrategy() schema.Namer
}
