package main

import (
	"context"

	"github.com/gotomicro/ego"
	"github.com/gotomicro/ego/core/elog"

	"github.com/ego-component/egorm"
)

// 运行方式：
// export EGO_DEBUG=true && go run main.go --config=config.toml
type KV struct {
	ID int    `gorm:"primaryKey" json:"id"`
	K  string `gorm:"uniqueIndex;not null" json:"k"`
	V  string `gorm:"not null" json:"v"`
}

func (KV) TableName() string { return "kv" }

var dbs []*egorm.Component

func main() {
	if err := ego.New().Invoker(
		openDB,
		run,
	).Run(); err != nil {
		elog.Error("startup", elog.Any("err", err))
	}
}

func openDB() error {
	dbs = []*egorm.Component{
		egorm.Load("sqlite.test").Build(),
		egorm.Load("sqlite.share.memory").Build(),
		egorm.Load("sqlite.nonshared.memory").Build(),
	}
	for _, db := range dbs {
		if err := db.AutoMigrate(&KV{}); err != nil {
			return err
		}
	}
	return nil
}

func run() error {
	ctx := context.Background()
	for _, db := range dbs {
		_ = db.WithContext(ctx).Create(&KV{K: "hello", V: "world"}).Error
		var out KV
		err := db.WithContext(ctx).Where("k = ?", "hello").First(&out).Error
		elog.Info("kv", elog.String("k", out.K), elog.String("v", out.V), elog.FieldErr(err))
	}
	return nil
}
