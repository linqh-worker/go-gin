package util

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"
)

// ParseConfigValue 校验并转换配置值；整数限制在JavaScript安全范围内。
func ParseConfigValue(kind, value string) (any, error) {
	switch kind {
	case "string":
		return value, nil
	case "integer":
		n, e := strconv.ParseInt(value, 10, 64)
		if e != nil || n > 9007199254740991 || n < -9007199254740991 {
			return nil, errors.New("请输入有效的安全整数")
		}
		return n, nil
	case "boolean":
		if value != "true" && value != "false" {
			return nil, errors.New("布尔值须为true或false")
		}
		return value == "true", nil
	case "json":
		if !json.Valid([]byte(value)) {
			return nil, errors.New("JSON格式错误")
		}
		var out any
		d := json.NewDecoder(strings.NewReader(value))
		d.UseNumber()
		if e := d.Decode(&out); e != nil {
			return nil, e
		}
		return out, nil
	default:
		return nil, errors.New("不支持的配置数据类型")
	}
}
