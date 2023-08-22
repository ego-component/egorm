package egorm

import (
	"log"
	"strings"
	"time"

	"github.com/spf13/cast"
	"gorm.io/gorm"
)

type (
	// Cond 为字段查询结构体
	Cond struct {
		// Connector MySQL中查询条件连接符，如and,or，默认为and
		Connector string
		// Op MySQL中查询条件，如like,=,in
		Op string
		// Val 查询条件对应的值
		Val interface{}
	}

	// Conds 为Cond类型map，用于定义Where方法参数 map[field.name]interface{}
	Conds map[string]interface{}

	// Ups 为更新某一条记录时存放的变更数据集合 map[field.name]field.value
	//Ups = map[string]interface{}
)

// assertCond 断言cond基本类型并返回Cond
// 如果是基本类型，则Cond.Op为"="
// 如果是切片类型，则Cond.Op为"in"。NOTICE: 不支持自定义类型切片，比如 type IDs []int
func assertCond(cond interface{}) Cond {
	// 先尝试断言为基本类型
	switch v := cond.(type) {
	case Cond:
		// 如果没有设置Connector，则默认为AND
		if v.Connector == "" {
			v.Connector = "AND"
		}
		return v
	case string:
		return Cond{"AND", "=", v}
	case bool:
		return Cond{"AND", "=", v}
	case float64:
		return Cond{"AND", "=", v}
	case float32:
		return Cond{"AND", "=", v}
	case int:
		return Cond{"AND", "=", v}
	case int64:
		return Cond{"AND", "=", v}
	case int32:
		return Cond{"AND", "=", v}
	case int16:
		return Cond{"AND", "=", v}
	case int8:
		return Cond{"AND", "=", v}
	case uint:
		return Cond{"AND", "=", v}
	case uint64:
		return Cond{"AND", "=", v}
	case uint32:
		return Cond{"AND", "=", v}
	case uint16:
		return Cond{"AND", "=", v}
	case uint8:
		return Cond{"AND", "=", v}
	case time.Duration:
		return Cond{"AND", "=", v}
	}

	// 再尝试断言为stringSlice类型
	condValueStr, err := cast.ToStringSliceE(cond)
	if err == nil {
		return Cond{"AND", "in", condValueStr}
	}

	// 再尝试断言为intSlice类型
	condValueInt, err := cast.ToIntSliceE(cond)
	if err == nil {
		return Cond{"AND", "in", condValueInt}
	}

	// 未识别的类型
	log.Printf("[assertCond] unrecognized type fail,%+v\n", cond)
	return Cond{}
}

// BuildQuery 根据conds构建sql和绑定的参数
func BuildQuery(conds Conds) (sql string, binds []interface{}) {
	sql = "1=1"
	binds = make([]interface{}, 0, len(conds))
	for field, cond := range conds {
		condVal := assertCond(cond)

		// 说明有表的数据
		if strings.Contains(field, ".") {
			arr := strings.Split(field, ".")
			if len(arr) != 2 {
				return
			}
			field = "`" + arr[0] + "`.`" + arr[1] + "`"
		} else {
			field = "`" + field + "`"
		}

		switch strings.ToLower(condVal.Op) {
		case "like":
			if condVal.Val != "" {
				sql += " " + condVal.Connector + " " + field + " like ?"
				condVal.Val = "%" + condVal.Val.(string) + "%"
			}
		case "%like":
			if condVal.Val != "" {
				sql += " " + condVal.Connector + " " + field + " like ?"
				condVal.Val = "%" + condVal.Val.(string)
			}
		case "like%":
			if condVal.Val != "" {
				sql += " " + condVal.Connector + " " + field + " like ?"
				condVal.Val = condVal.Val.(string) + "%"
			}
		case "in", "not in":
			sql += " " + condVal.Connector + " " + field + condVal.Op + " (?) "
		case "between":
			sql += " " + condVal.Connector + " " + field + condVal.Op + " ? AND ?"
			val := cast.ToStringSlice(condVal.Val)
			binds = append(binds, val[0], val[1])
			continue
		case "exp":
			sql += " " + condVal.Connector + " " + field + " ? "
			condVal.Val = gorm.Expr(condVal.Val.(string))
		default:
			sql += " " + condVal.Connector + " " + field + condVal.Op + " ? "
		}
		binds = append(binds, condVal.Val)
	}
	return
}
