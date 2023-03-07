package egorm

import (
	"crypto/tls"

	"github.com/gotomicro/ego/core/elog"
)

type Authentication struct {
	// TLS authentication
	TLS *TLSConfig
}

func (config *Authentication) TLSConfig() *tls.Config {
	if config.TLS != nil {
		tlsConfig, err := config.TLS.LoadTLSConfig()
		if err != nil {
			elog.Panic("error loading tls config", elog.FieldErr(err))
			return nil
		}
		if tlsConfig != nil && tlsConfig.InsecureSkipVerify {
			return tlsConfig
		}
	}
	return nil
}
