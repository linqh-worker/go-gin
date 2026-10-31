package user

import "time"

type AuthUser struct {
	ID          int        `json:"id"`
	Username    string     `json:"username"`
	Nickname    string     `json:"nickname"`
	AvatarURL   string     `json:"avatar_url"`
	Phone       *string    `json:"phone"`
	Email       *string    `json:"email"`
	Gender      uint8      `json:"gender"`
	LastLoginAt *time.Time `json:"last_login_at"`
	LastLoginIP string     `json:"last_login_ip"`
	LoginCount  int64      `json:"login_count"`
	UpdatedAt   *time.Time `json:"updated_at"`
	Age         *int       `json:"age"`
	Status      int        `json:"status"`
	CreatedAt   *time.Time `json:"created_at"`
}
