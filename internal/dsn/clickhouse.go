package dsn

import (
	"errors"
	"net/url"

	"gorm.io/driver/clickhouse"
	"gorm.io/gorm"

	"github.com/ego-component/egorm/manager"
)

var (
	_             manager.DSNParser = (*ClickHouseDSNParser)(nil)
	errInvalidDSN                   = errors.New("invalid dsn")
)

type ClickHouseDSNParser struct {
}

func init() {
	manager.Register(&ClickHouseDSNParser{})
}

func (p *ClickHouseDSNParser) Scheme() string {
	return "clickhouse"
}

func (p *ClickHouseDSNParser) GetDialector(dsn string) gorm.Dialector {
	return clickhouse.Open(dsn)
}

func (p *ClickHouseDSNParser) ParseDSN(dsn string) (cfg *manager.DSN, err error) {
	cfg = new(manager.DSN)
	u, err := url.Parse(dsn)
	if err != nil || u.Host == "" {
		return nil, errInvalidDSN
	}
	cfg.Addr = u.Host
	cfg.Net = u.Scheme
	params := u.Query()
	cfg.Params = make(map[string]string, 0)
	for key, param := range params {
		switch key {
		case "username":
			cfg.User = param[0]
		case "password":
			cfg.Password = param[0]
		case "database":
			cfg.DBName = param[0]
		default:
			cfg.Params[key] = param[0]
		}
	}
	return cfg, nil
}
