package egorm

import (
	"time"

	"gorm.io/gorm/schema"

	"github.com/ego-component/egorm/manager"
)

// Option 可选项
type Option func(c *Container)

// WithDSN 设置dsn
func WithDSN(dsn string) Option {
	return func(c *Container) {
		c.config.DSN = dsn
	}
}

// WithDialect 设置 dialect
func WithDialect(dialect string) Option {
	return func(c *Container) {
		c.config.Dialect = dialect
	}
}

// WithDSNParser 设置自定义dsnParser
func WithDSNParser(parser manager.DSNParser) Option {
	return func(c *Container) {
		c.dsnParser = parser
	}
}

func WithNamingStrategy(namingStrategy schema.Namer) Option {
	return func(c *Container) {
		c.config.namingStrategy = namingStrategy
	}
}

// WithInterceptor 设置自定义拦截器
func WithInterceptor(is ...Interceptor) Option {
	return func(c *Container) {
		if c.config.interceptors == nil {
			c.config.interceptors = make([]Interceptor, 0)
		}
		c.config.interceptors = append(c.config.interceptors, is...)
	}
}

func WithEnableAccessInterceptor(enableAccessInterceptor bool) Option {
	return func(c *Container) {
		c.config.EnableAccessInterceptor = enableAccessInterceptor
	}
}

func WithEnableAccessInterceptorReq(enableAccessInterceptorReq bool) Option {
	return func(c *Container) {
		c.config.EnableAccessInterceptorReq = enableAccessInterceptorReq
	}
}

func WithEnableAccessInterceptorRes(enableAccessInterceptorRes bool) Option {
	return func(c *Container) {
		c.config.EnableAccessInterceptorRes = enableAccessInterceptorRes
	}
}

func WithMaxIdleConns(maxIdleConns int) Option {
	return func(c *Container) {
		c.config.MaxIdleConns = maxIdleConns
	}
}

func WithMaxOpenConns(maxOpenConns int) Option {
	return func(c *Container) {
		c.config.MaxOpenConns = maxOpenConns
	}
}

func WithConnMaxLifetime(connMaxLifetime time.Duration) Option {
	return func(c *Container) {
		c.config.ConnMaxLifetime = connMaxLifetime
	}
}

func WithOnFail(onFail string) Option {
	return func(c *Container) {
		c.config.OnFail = onFail
	}
}

func WithSlowLogThreshold(slowLogThreshold time.Duration) Option {
	return func(c *Container) {
		c.config.SlowLogThreshold = slowLogThreshold
	}
}

func WithEnableDetailSQL(enableDetailSQL bool) Option {
	return func(c *Container) {
		c.config.EnableDetailSQL = enableDetailSQL
	}
}
