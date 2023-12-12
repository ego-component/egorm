package egorm

import (
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
