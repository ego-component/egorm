package egorm

import (
	"net/http"
	"sync"

	"github.com/gotomicro/ego/core/elog"
	"github.com/gotomicro/ego/core/emetric"
	"github.com/gotomicro/ego/server/egovernor"
	jsoniter "github.com/json-iterator/go"
)

var instances = sync.Map{}

func init() {
	egovernor.HandleFunc("/debug/gorm/stats", func(w http.ResponseWriter, r *http.Request) {
		_ = jsoniter.NewEncoder(w).Encode(stats())
	})
	go monitor()
}

func monitor() {
	for {
		instances.Range(func(key, val interface{}) bool {
			name := key.(string)
			db := val.(*Component)

			sqlDB, err := db.DB()
			if err != nil {
				elog.EgoLogger.With(elog.FieldComponent(PackageName)).Panic("monitor db error", elog.FieldErr(err))
				return false
			}
			stats := sqlDB.Stats()
			emetric.ClientStatsGauge.Set(float64(stats.MaxOpenConnections), emetric.TypeGorm, name, "max_open_connections")
			emetric.ClientStatsGauge.Set(float64(stats.OpenConnections), emetric.TypeGorm, name, "open_connections")
			emetric.ClientStatsGauge.Set(float64(stats.InUse), emetric.TypeGorm, name, "in_use")
			emetric.ClientStatsGauge.Set(float64(stats.Idle), emetric.TypeGorm, name, "idle")
			emetric.ClientStatsGauge.Set(float64(stats.WaitCount), emetric.TypeGorm, name, "wait_count")
			emetric.ClientStatsGauge.Set(float64(stats.WaitDuration), emetric.TypeGorm, name, "wait_duration")
			emetric.ClientStatsGauge.Set(float64(stats.MaxIdleClosed), emetric.TypeGorm, name, "max_idle_closed")
			emetric.ClientStatsGauge.Set(float64(stats.MaxIdleTimeClosed), emetric.TypeGorm, name, "max_idle_time_closed")
			emetric.ClientStatsGauge.Set(float64(stats.MaxLifetimeClosed), emetric.TypeGorm, name, "max_lifetime_closed")
			return true
		})
	}
}

// stats
func stats() (stats map[string]interface{}) {
	stats = make(map[string]interface{})
	instances.Range(func(key, val interface{}) bool {
		name := key.(string)
		db := val.(*Component)

		sqlDB, err := db.DB()
		if err != nil {
			elog.EgoLogger.With(elog.FieldComponent(PackageName)).Panic("stats db error", elog.FieldErr(err))
			return false
		}
		stats[name] = sqlDB.Stats()
		return true
	})

	return
}
