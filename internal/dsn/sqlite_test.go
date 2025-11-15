package dsn

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSqliteDSNParser_ParseDSN(t *testing.T) {
	tests := []struct {
		name     string
		dsn      string
		expected *struct {
			DBName string
			Addr   string
			Params map[string]string
		}
	}{
		{
			name: "memory database",
			dsn:  ":memory:",
			expected: &struct {
				DBName string
				Addr   string
				Params map[string]string
			}{
				DBName: "memory",
				Addr:   "",
				Params: map[string]string{},
			},
		},
		{
			name: "file memory with cache",
			dsn:  "file::memory:?cache=shared",
			expected: &struct {
				DBName string
				Addr   string
				Params map[string]string
			}{
				DBName: "memory",
				Addr:   "",
				Params: map[string]string{
					"cache": "shared",
				},
			},
		},
		{
			name: "file based database",
			dsn:  "test.db",
			expected: &struct {
				DBName string
				Addr   string
				Params map[string]string
			}{
				DBName: "test.db",
				Addr:   "",
				Params: map[string]string{},
			},
		},
		{
			name: "file URI with parameters",
			dsn:  "file:./data/app.db?cache=shared&_journal_mode=WAL&_busy_timeout=5000&_fk=1",
			expected: &struct {
				DBName string
				Addr   string
				Params map[string]string
			}{
				DBName: "app.db",
				Addr:   "",
				Params: map[string]string{
					"cache":         "shared",
					"_journal_mode": "WAL",
					"_busy_timeout": "5000",
					"_fk":           "1",
				},
			},
		},
	}

	dsnParser := SqliteDSNParser{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := dsnParser.ParseDSN(tt.dsn)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected.DBName, cfg.DBName)
			assert.Equal(t, tt.expected.Addr, cfg.Addr)
			assert.Equal(t, tt.expected.Params, cfg.Params)
		})
	}
}
