package user

import "time"

type AuthUser struct {
	ID          int     `json:"id"`
	Username    string  `json:"username"`
	Nickname    string  `json:"nickname"`
	AvatarURL   string  `json:"avatar_url"`
	Phone       *string `json:"phone"`
	Email       *string `json:"email"`
	Gender      uint8   `json:"gender"`
	LastLoginAt *string `json:"last_login_at"`
	LastLoginIP string  `json:"last_login_ip"`
	LoginCount  int64   `json:"login_count"`
	UpdatedAt   *string `json:"updated_at"`
	Age         *int    `json:"age"`
	Status      int     `json:"status"`
	CreatedAt   *string `json:"created_at"`
}

// FormatTime 统一用户响应时间为北京时间；数据库空值保持 JSON null。
func FormatTime(value *time.Time) *string {
	if value == nil {
		return nil
	}
	text := value.In(time.FixedZone("UTC+8", 8*60*60)).Format("2006-01-02 15:04:05")
	return &text
}
