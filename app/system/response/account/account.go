package account

type Profile struct {
	ID          int     `json:"id"`
	Username    string  `json:"username"`
	Nickname    string  `json:"nickname"`
	AvatarURL   string  `json:"avatar_url"`
	Phone       *string `json:"phone"`
	Email       *string `json:"email"`
	Age         *int    `json:"age"`
	Gender      uint8   `json:"gender"`
	Status      int     `json:"status"`
	LoginCount  int64   `json:"login_count"`
	LastLoginAt *string `json:"last_login_at"`
	CreatedAt   *string `json:"created_at"`
	UpdatedAt   *string `json:"updated_at"`
}
type Page struct {
	Items    []Profile `json:"items"`
	Total    int64     `json:"total"`
	Page     int       `json:"page"`
	PageSize int       `json:"page_size"`
}
