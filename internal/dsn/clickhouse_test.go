package dsn

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClickhouseDsnParser_ParseDSN(t *testing.T) {
	dsn := "tcp://localhost:9000?database=dbname&username=user&password=password&read_timeout=10&write_timeout=20&foo"
	var parser ClickHouseDSNParser
	cfg, err := parser.ParseDSN(dsn)
	assert.NoError(t, err)
	assert.Equal(t, "user", cfg.User)
	assert.Equal(t, "password", cfg.Password)
	assert.Equal(t, "dbname", cfg.DBName)
	assert.Equal(t, "localhost:9000", cfg.Addr)
	assert.Equal(t, "10", cfg.Params["read_timeout"])
	assert.Equal(t, "20", cfg.Params["write_timeout"])
	assert.Equal(t, "tcp", cfg.Net)
	assert.Equal(t, "", cfg.Params["foo"])
	_, err = parser.ParseDSN("some-wrong-dsn")
	assert.ErrorIs(t, err, errInvalidDSN)
}
