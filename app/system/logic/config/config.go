package config

import (
	"context"
	"go-gin/app/system/dao"
	request "go-gin/app/system/request/config"
	response "go-gin/app/system/response/config"
	"go-gin/internal/errorx"
	"go-gin/util"
	"regexp"
	"strings"
	"unicode/utf8"
)

func Groups(ctx context.Context) ([]response.Group, error) {
	rows, err := dao.ConfigItems(ctx)
	if err != nil {
		return nil, err
	}
	groups := []response.Group{}
	indexes := map[string]int{}
	appendGroup := func(r dao.ConfigItem) {
		name := r.GroupName
		if name == "" {
			name = r.GroupCode
		}
		indexes[r.GroupCode] = len(groups)
		groups = append(groups, response.Group{Code: r.GroupCode, Name: name, Version: r.Version, IsBuiltin: r.IsBuiltin, Items: []response.Item{}})
	}
	for _, r := range rows {
		if r.RecordType == "group" && r.GroupCode == "website" {
			appendGroup(r)
		}
	}
	for _, r := range rows {
		if r.RecordType == "group" && r.GroupCode != "website" {
			appendGroup(r)
		}
	}
	for _, r := range rows {
		if r.RecordType != "item" {
			continue
		}
		i, ok := indexes[r.GroupCode]
		if !ok {
			continue
		}
		groups[i].Items = append(groups[i].Items, response.Item{ID: r.ID, Sort: r.Sort, Key: r.ConfigKey, Name: r.Name, Value: r.ConfigValue, ValueType: r.ValueType, InputType: r.InputType, IsPublic: r.IsPublic, IsBuiltin: r.IsBuiltin, Remark: r.Remark, Version: r.Version, UpdatedAt: util.FormatTime(r.UpdatedAt)})
	}
	return groups, nil
}
func Save(ctx context.Context, group string, actor int, r request.SaveReq) error {
	if actor <= 0 {
		return errorx.New(401, "请先登录")
	}
	if !regexp.MustCompile(`^[a-z][a-z0-9_]{0,63}$`).MatchString(group) {
		return errorx.New(10001, "分组编码无效")
	}
	if len(r.Items) == 0 || len(r.Items) > 1000 {
		return errorx.New(10001, "每次保存1至1000项配置")
	}
	seen := map[string]bool{}
	changes := make([]dao.ConfigChange, 0, len(r.Items))
	for _, v := range r.Items {
		if v.Key == "" || seen[v.Key] {
			return errorx.New(10001, "配置键不能为空或重复")
		}
		seen[v.Key] = true
		if len(v.Value) > 16000 {
			return errorx.New(10001, "单项配置不得超过16000字节")
		}
		changes = append(changes, dao.ConfigChange{Key: v.Key, Value: v.Value, Version: v.Version})
	}
	return dao.SaveConfigGroup(ctx, group, actor, changes, validateValue)
}
func validateValue(item dao.ConfigItem, value string) error {
	if _, err := util.ParseConfigValue(item.ValueType, value); err != nil {
		return errorx.New(10001, item.Name+"："+err.Error())
	}
	max := 0
	switch item.ConfigKey {
	case "website.name":
		max = 100
	case "website.home_title":
		max = 200
	case "website.keywords":
		max = 500
	case "website.description":
		max = 1000
	}
	if max > 0 && utf8.RuneCountInString(value) > max {
		return errorx.New(10001, item.Name+"长度超过限制")
	}
	if (item.ConfigKey == "website.name" || item.ConfigKey == "website.home_title") && strings.TrimSpace(value) == "" {
		return errorx.New(10001, item.Name+"不能为空")
	}
	return nil
}

func validGroup(code string) bool {
	return regexp.MustCompile(`^[a-z][a-z0-9_]{0,63}$`).MatchString(code)
}
func CreateGroup(ctx context.Context, actor int, r request.GroupReq) error {
	if !validGroup(r.Code) || strings.TrimSpace(r.Name) == "" {
		return errorx.New(10001, "请填写有效的分组编码和名称")
	}
	return dao.CreateConfigGroup(ctx, r.Code, strings.TrimSpace(r.Name), actor)
}
func UpdateGroup(ctx context.Context, code string, actor int, r request.GroupReq) error {
	if !validGroup(code) || strings.TrimSpace(r.Name) == "" || r.Version < 1 {
		return errorx.New(10001, "分组名称或版本无效")
	}
	if r.Code != "" && r.Code != code {
		return errorx.New(10001, "分组编码创建后不能修改")
	}
	return dao.UpdateConfigGroup(ctx, code, strings.TrimSpace(r.Name), r.Version, actor)
}
func DeleteGroup(ctx context.Context, code string, version int64) error {
	if !validGroup(code) || version < 1 {
		return errorx.New(10001, "分组或版本无效")
	}
	return dao.DeleteConfigGroup(ctx, code, version)
}
func SaveDefinition(ctx context.Context, id int64, actor int, r request.DefinitionReq) error {
	if !validGroup(r.GroupCode) || !regexp.MustCompile(`^[a-z][a-z0-9_]*(\.[a-z][a-z0-9_]*)*$`).MatchString(r.Key) || len(r.Key) > 128 {
		return errorx.New(10001, "key须为小写字母开头，支持数字、下划线和点分段")
	}
	if strings.TrimSpace(r.Name) == "" || len(r.Value) > 16000 || id > 0 && r.Version < 1 {
		return errorx.New(10001, "配置名称、值大小或版本无效")
	}
	compatible := r.ValueType == "string" && (r.InputType == "input" || r.InputType == "textarea") || r.ValueType == "integer" && r.InputType == "number" || r.ValueType == "boolean" && r.InputType == "switch" || r.ValueType == "json" && r.InputType == "json"
	if !compatible {
		return errorx.New(10001, "数据类型与输入控件不匹配")
	}
	item := dao.ConfigItem{GroupCode: r.GroupCode, ConfigKey: r.Key, Name: strings.TrimSpace(r.Name), ConfigValue: r.Value, ValueType: r.ValueType, InputType: r.InputType, IsPublic: r.IsPublic, IsBuiltin: r.IsBuiltin, Sort: r.Sort, Remark: r.Remark, Version: r.Version}
	return dao.SaveConfigDefinition(ctx, id, item, actor, validateValue)
}
func DeleteDefinition(ctx context.Context, id, version int64, actor int) error {
	if id < 1 || version < 1 {
		return errorx.New(10001, "ID或版本无效")
	}
	return dao.DeleteConfigDefinition(ctx, id, version, actor)
}
