package util

import "time"

var beijingZone = time.FixedZone("UTC+8", 8*60*60)

// FormatTime 统一响应时间为北京时间；数据库空值保持 JSON null。
func FormatTime(value *time.Time) *string {
	if value == nil {
		return nil
	}
	text := value.In(beijingZone).Format("2006-01-02 15:04:05")
	return &text
}
