# egorm 组件使用指南
## 1 简介
对 [gorm](https://github.com/go-gorm/gorm) 进行了轻量封装，并提供了以下功能：
- 规范了标准配置格式，提供了统一的 Load().Build() 方法。
- 支持自定义拦截器
- 提供了默认的 Debug 拦截器，开启 Debug 后可输出 Request、Response 至终端。
- 提供了默认的 Metric 拦截器，开启后可采集 Prometheus 指标数据
- 提供了默认的 OpenTelemetry 拦截器，开启后可采集 Tracing Span 数据

## 2 说明
* [example地址](https://github.com/ego-component/egorm/tree/master/examples)
* [文档地址](https://ego.gocn.vip/frame/client/gorm.html#_1-%E7%AE%80%E4%BB%8B)
* ego版本：``ego@v1.0.0``
* egorm版本: ``egorm@1.0.0``

## 3 使用方式
```bash
go get github.com/ego-component/egorm
```

## 4 GORM配置
```go
type Config struct {
    Dialect                    string        // 选择数据库种类，默认mysql,postgres,mssql
    DSN                        string        // DSN地址: mysql://username:password@tcp(127.0.0.1:3306)/mysql?charset=utf8mb4&collation=utf8mb4_general_ci&parseTime=True&loc=Local&timeout=1s&readTimeout=3s&writeTimeout=3s
    Debug                      bool          // 是否开启调试，默认不开启，开启后并加上export EGO_DEBUG=true，可以看到每次请求，配置名、地址、耗时、请求数据、响应数据
    RawDebug                   bool          // 是否开启原生调试开关，默认不开启
    MaxIdleConns               int           // 最大空闲连接数，默认10
    MaxOpenConns               int           // 最大活动连接数，默认100
    ConnMaxLifetime            time.Duration // 连接的最大存活时间，默认300s
    OnFail                     string        // 创建连接的错误级别，=panic时，如果创建失败，立即panic，默认连接不上panic
    SlowLogThreshold           time.Duration // 慢日志阈值，默认500ms
    EnableMetricInterceptor    bool          // 是否开启监控，默认开启
    EnableTraceInterceptor     bool          // 是否开启链路追踪，默认开启
    EnableDetailSQL            bool          // 记录错误sql时,是否打印包含参数的完整sql语句，select * from aid = ?;
    EnableAccessInterceptor    bool          // 是否开启，记录请求数据
    EnableAccessInterceptorReq bool          // 是否开启记录请求参数
    EnableAccessInterceptorRes bool          // 是否开启记录响应参数
}
```

## 5 普通GORM查询
## 5.1 用户配置
```toml
[mysql.test]
   debug = true # ego重写gorm debug，打开后可以看到，配置名、代码执行行号、地址、耗时、请求数据、响应数据
   dsn = "root:root@tcp(127.0.0.1:3306)/ego?charset=utf8&parseTime=True&loc=Local&readTimeout=1s&timeout=1s&writeTimeout=3s"
```

## 5.2 优雅的Debug
通过开启``debug``配置和命令行的``export EGO_DEBUG=true``，我们就可以在测试环境里看到请求里的配置名、地址、耗时、请求数据、响应数据
![image](https://cdn.gocn.vip/ego/assets/img/ego_debug.4672a95e.png)
当然你也可以开启``gorm``原生的调试，将``rawDebug``设置为``true``

## 5.3 用户代码
配置创建一个 ``gorm`` 的配置项，其中内容按照上文配置进行填写。以上这个示例里这个配置key是``gorm.test``

代码中创建一个 ``gorm`` 实例 ``egorm.Load("key").Build()``，代码中的 ``key`` 和配置中的 ``key`` 要保持一致。创建完 ``gorm`` 实例后，就可以直接使用他对 ``db`` 进行 ``crud`` 。

```go
package main

import (
	"github.com/gotomicro/ego"
	"github.com/ego-component/egorm"
	"github.com/gotomicro/ego/core/elog"
)

/**
1.新建一个数据库叫test
2.执行以下example，export EGO_DEBUG=true && go run main.go --config=config.toml
*/
type User struct {
	Id       int    `gorm:"not null" json:"id"`
	Nickname string `gorm:"not null" json:"name"`
}

func (User) TableName() string {
	return "user2"
}

func main() {
	err := ego.New().Invoker(
		openDB,
		testDB,
	).Run()
	if err != nil {
		elog.Error("startup", elog.Any("err", err))
	}
}

var gormDB *egorm.Component

func openDB() error {
	gormDB = egorm.Load("mysql.test").Build()
	models := []interface{}{
		&User{},
	}
	gormDB.SingularTable(true)
	gormDB.Set("gorm:table_options", "ENGINE=InnoDB").AutoMigrate(models...)
	gormDB.Create(&User{
		Nickname: "ego",
	})
	return nil
}

func testDB() error {
	var user User
	err := gormDB.Where("id = ?", 100).Find(&user).Error
	elog.Info("user info", elog.String("name", user.Nickname))
	return err
}
```
## 6 GORM的日志
任何gorm的请求都会记录gorm的错误access日志，如果需要对gorm的日志做定制化处理，可参考以下使用方式。

### 6.1 开启GORM的access日志
线上在并发量不高，或者核心业务下可以开启全量access日志，这样方便我们排查问题

### 6.2 开启日志方式
在原有的mysql配置中，加入以下三行配置，gorm的日志里就会记录响应的数据
```toml
[mysql.test]
enableAccessInterceptor=true       # 是否开启，记录请求数据
enableAccessInterceptorReq=true    # 是否开启记录请求参数
enableAccessInterceptorRes=true    # 是否开启记录响应参数
```

![img.png](https://cdn.gocn.vip/ego/assets/img/enable_req_res.89087f89.png)

### 6.3 开启日志的详细数据
记录请求参数日志为了安全起见，默认是不开启详细的sql数据，记录的是参数绑定的SQL日志，如果需要开启详细数据，需要在配置里添加``
```toml
[mysql.test]
enableDetailSQL=true       # 记录sql时,是否打印包含参数的完整sql语句，select * from aid = ?;
```
![img.png](https://cdn.gocn.vip/ego/assets/img/enable_req_res_detail.c932d5dc.png)

### 6.4 开启日志的链路数据
代码方面使用`db.WithContext(ctx)`，会在access日志中自动记录trace id信息

### 6.5 开启自定义日志字段的数据
在使用了ego的自定义字段功能`export EGO_LOG_EXTRA_KEYS=X-Ego-Uid`，将对应的数据塞入到context中，那么gorm的access日志就可以记录对应字段信息。
参考[详细文档](https://ego.gocn.vip/micro/chapter2/trace.html#_6-ego-access-%E8%87%AA%E5%AE%9A%E4%B9%89%E9%93%BE%E8%B7%AF)：
```go
func testDB() error {
	var user User
	for _, db := range DBs {
		ctx := context.Background()
		ctx = transport.WithValue(ctx, "X-Ego-Uid", 9527)
		err := db.WithContext(ctx).Where("id = ?", 100).First(&user).Error
		elog.Info("user info", elog.String("name", user.Nickname), elog.FieldErr(err))
	}
	return nil
}
```