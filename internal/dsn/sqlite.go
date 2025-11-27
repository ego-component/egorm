package dsn

import (
	"net/url"
	"path/filepath"
	"strings"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"

	"github.com/ego-component/egorm/manager"
)

var (
	_ manager.DSNParser = (*SqliteDSNParser)(nil)
)

type SqliteDSNParser struct {
}

func init() {
	manager.Register(&SqliteDSNParser{})
}

func (p *SqliteDSNParser) Scheme() string {
	return "sqlite"
}

func (p *SqliteDSNParser) NamingStrategy() schema.Namer {
	return nil
}

func (p *SqliteDSNParser) GetDialector(dsn string) gorm.Dialector {
	return sqlite.Open(dsn)
}

// ParseDSN supports typical gorm sqlite DSN strings:
// - file based: "test.db", "/abs/path/to/test.db", "file:test.db?cache=shared&_fk=1"
// - memory: ":memory:", "file::memory:?cache=shared"
func (p *SqliteDSNParser) ParseDSN(dsn string) (cfg *manager.DSN, err error) {
	cfg = new(manager.DSN)
	cfg.Params = map[string]string{}

	// Normalize for parsing query parameters
	raw := dsn
	if strings.HasPrefix(raw, "file:") {
		if idx := strings.IndexByte(raw, '?'); idx >= 0 {
			query := raw[idx+1:]
			for _, kv := range strings.Split(query, "&") {
				parts := strings.SplitN(kv, "=", 2)
				if len(parts) != 2 {
					continue
				}
				val, decodeErr := url.QueryUnescape(parts[1])
				if decodeErr != nil {
					continue
				}
				cfg.Params[parts[0]] = val
			}
		}
	}

	// Determine DBName for logging/metrics only
	switch {
	// memory mode
	case dsn == ":memory:":
	case strings.Contains(dsn, "::memory:"):
		cfg.DBName = "memory"
	default:
		trimmed := dsn
		// strip "file:" prefix for name extraction
		if strings.HasPrefix(trimmed, "file:") {
			if idx := strings.IndexByte(trimmed, '?'); idx >= 0 {
				trimmed = trimmed[:idx]
			}
			trimmed = strings.TrimPrefix(trimmed, "file:")
		} else {
			// if path like "dir/db.sqlite" keep base
			if idx := strings.IndexByte(trimmed, '?'); idx >= 0 {
				trimmed = trimmed[:idx]
			}
		}
		base := filepath.Base(trimmed)
		if base == "" || base == "." || base == "/" {
			base = "sqlite"
		}
		cfg.DBName = base
	}
	return
}
