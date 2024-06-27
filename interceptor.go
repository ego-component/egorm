package egorm

import (
	"context"
	"errors"
	"fmt"
	"log"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/gotomicro/ego/core/eapp"
	"github.com/gotomicro/ego/core/elog"
	"github.com/gotomicro/ego/core/emetric"
	"github.com/gotomicro/ego/core/etrace"
	"github.com/gotomicro/ego/core/transport"
	"github.com/gotomicro/ego/core/util/xdebug"
	"github.com/spf13/cast"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"
	"gorm.io/hints"

	"github.com/ego-component/egorm/manager"
)

// Handler ...
type Handler func(*gorm.DB)

// Processor ...
type Processor interface {
	Get(name string) func(*gorm.DB)
	Replace(name string, handler func(*gorm.DB)) error
}

// Interceptor ...
type Interceptor func(string, *manager.DSN, string, *config, *elog.Component) func(next Handler) Handler

func debugInterceptor(compName string, dsn *manager.DSN, _ string, _ *config, _ *elog.Component) func(Handler) Handler {
	return func(next Handler) Handler {
		return func(db *gorm.DB) {
			if !eapp.IsDevelopmentMode() {
				next(db)
				return
			}
			beg := time.Now()
			next(db)
			cost := time.Since(beg)
			if db.Error != nil {
				log.Println("[egorm.response]",
					xdebug.MakeReqAndResError(fileWithLineNum(), compName, fmt.Sprintf("%v", dsn.Addr+"/"+dsn.DBName), cost, logSQL(db, true), db.Error.Error()),
				)
			} else {
				log.Println("[egorm.response]",
					xdebug.MakeReqAndResInfo(fileWithLineNum(), compName, fmt.Sprintf("%v", dsn.Addr+"/"+dsn.DBName), cost, logSQL(db, true), fmt.Sprintf("%v", db.Statement.Dest)),
				)
			}

		}
	}
}

func metricInterceptor(compName string, dsn *manager.DSN, op string, config *config, logger *elog.Component) func(Handler) Handler {
	return func(next Handler) Handler {
		return func(db *gorm.DB) {
			beg := time.Now()
			next(db)
			cost := time.Since(beg)

			var fields = make([]elog.Field, 0, 15+transport.CustomContextKeysLength())
			event := "normal"
			fields = append(fields,
				elog.FieldMethod(op),
				elog.FieldName(dsn.DBName+"."+db.Statement.Table),
				elog.FieldCost(cost),
			)
			if config.EnableAccessInterceptorReq {
				fields = append(fields, elog.String("req", logSQL(db, config.EnableDetailSQL)))
			}
			if config.EnableAccessInterceptorRes {
				fields = append(fields, elog.Any("res", db.Statement.Dest))
			}
			// 开启了链路，那么就记录链路id
			if etrace.IsGlobalTracerRegistered() {
				fields = append(fields, elog.FieldTid(etrace.ExtractTraceID(db.Statement.Context)))
			}

			// 支持自定义log
			for _, key := range transport.CustomContextKeys() {
				if value := getContextValue(db.Statement.Context, key); value != "" {
					fields = append(fields, elog.FieldCustomKeyValue(key, value))
				}
			}

			// 记录监控耗时
			emetric.ClientHandleHistogram.WithLabelValues(emetric.TypeGorm, compName, dsn.DBName+"."+db.Statement.Table, dsn.Addr).Observe(cost.Seconds())
			isSlowLog := false
			if config.SlowLogThreshold > time.Duration(0) && config.SlowLogThreshold < cost {
				event = "slow"
				isSlowLog = true
			}

			// 如果有错误，记录错误信息
			if db.Error != nil {
				fields = append(fields,
					elog.FieldEvent(event),
					elog.FieldErr(db.Error),
				)
				if errors.Is(db.Error, ErrRecordNotFound) {
					// 这种日志可能很多，也没必要，只有开启的时候，或者慢日志的时候记录
					if config.EnableAccessInterceptor || isSlowLog {
						logger.Warn("access", fields...)
					}
					emetric.ClientHandleCounter.Inc(emetric.TypeGorm, compName, dsn.DBName+"."+db.Statement.Table, dsn.Addr, "Empty")
					return
				}
				// 如果用户没开启req，那么错误必记录Req
				if !config.EnableAccessInterceptorReq {
					fields = append(fields, elog.String("req", logSQL(db, true)))
				}
				logger.Error("access", fields...)
				emetric.ClientHandleCounter.Inc(emetric.TypeGorm, compName, dsn.DBName+"."+db.Statement.Table, dsn.Addr, "Error")
				return
			}

			emetric.ClientHandleCounter.Inc(emetric.TypeGorm, compName, dsn.DBName+"."+db.Statement.Table, dsn.Addr, "OK")

			if config.EnableAccessInterceptor || isSlowLog {
				fields = append(fields,
					elog.FieldEvent(event),
				)
				if isSlowLog {
					logger.Warn("access", fields...)
				} else {
					logger.Info("access", fields...)
				}
			}
		}
	}
}

func logSQL(db *gorm.DB, enableDetailSQL bool) string {
	if enableDetailSQL {
		return db.Explain(db.Statement.SQL.String(), db.Statement.Vars...)
	}
	return db.Statement.SQL.String()
}

func traceInterceptor(compName string, dsn *manager.DSN, _ string, options *config, _ *elog.Component) func(Handler) Handler {
	ip, port := peerInfo(dsn.Addr)
	attrs := []attribute.KeyValue{
		semconv.NetHostIPKey.String(ip),
		semconv.NetPeerPortKey.Int(port),
		semconv.NetTransportKey.String(dsn.Net),
		semconv.DBNameKey.String(dsn.DBName),
		attribute.String("db.component_name", compName),
	}
	tracer := etrace.NewTracer(trace.SpanKindClient)
	return func(next Handler) Handler {
		return func(db *gorm.DB) {
			if db.Statement.Context != nil {
				operation := "gorm:"
				if len(db.Statement.BuildClauses) > 0 {
					operation += strings.ToLower(db.Statement.BuildClauses[0])
				}
				_, span := tracer.Start(db.Statement.Context, operation, nil, trace.WithAttributes(attrs...))
				defer span.End()
				comment := fmt.Sprintf("tid=%s", span.SpanContext().TraceID().String())
				if db.Statement.SQL.Len() > 0 {
					sql := db.Statement.SQL.String()
					db.Statement.SQL.Reset()
					db.Statement.SQL.WriteString("/* ")
					db.Statement.SQL.WriteString(comment)
					db.Statement.SQL.WriteString(" */ ")
					db.Statement.SQL.WriteString(sql)
				} else {
					hints.CommentBefore("SELECT", comment).ModifyStatement(db.Statement)
					hints.CommentBefore("INSERT", comment).ModifyStatement(db.Statement)
					hints.CommentBefore("UPDATE", comment).ModifyStatement(db.Statement)
					hints.CommentBefore("DELETE", comment).ModifyStatement(db.Statement)
				}
				next(db)
				span.SetAttributes(
					semconv.DBSystemKey.String(db.Dialector.Name()),
					semconv.DBStatementKey.String(logSQL(db, options.EnableDetailSQL)),
					semconv.DBOperationKey.String(operation),
					semconv.DBSQLTableKey.String(db.Statement.Table),
					semconv.NetPeerNameKey.String(dsn.Addr),
					attribute.Int64("db.rows_affected", db.RowsAffected),
				)
				var err = db.Error
				if !options.TraceRecordErrorOnNotFound && errors.Is(err, gorm.ErrRecordNotFound) {
					err = nil
				}
				if err != nil {
					span.RecordError(db.Error)
					span.SetStatus(codes.Error, db.Error.Error())
					return
				}
				span.SetStatus(codes.Ok, "OK")
				return
			}
			next(db)
		}
	}
}

func getContextValue(c context.Context, key string) string {
	if key == "" {
		return ""
	}
	return cast.ToString(transport.Value(c, key))
}

// todo ipv6
func peerInfo(addr string) (hostname string, port int) {
	if idx := strings.IndexByte(addr, ':'); idx >= 0 {
		hostname = addr[:idx]
		port, _ = strconv.Atoi(addr[idx+1:])
	}
	return hostname, port
}

func fileWithLineNum() string {
	// the second caller usually from internal, so set i start from 2
	for i := 2; i < 15; i++ {
		_, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		if (!(strings.Contains(file, "ego-component/egorm") && strings.HasSuffix(file, "interceptor.go")) && !strings.Contains(file, "gorm.io/gorm")) || strings.HasSuffix(file, "_test.go") {
			return file + ":" + strconv.FormatInt(int64(line), 10)
		}
	}
	return ""
}
