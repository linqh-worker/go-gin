package account

type ListReq struct {
	Page     int    `form:"page" binding:"omitempty,min=1,max=1000000"`
	PageSize int    `form:"page_size" binding:"omitempty,min=1,max=100"`
	Keyword  string `form:"keyword" binding:"max=64"`
	Status   *int   `form:"status" binding:"omitempty,oneof=0 1"`
}
type CreateReq struct {
	AvatarURL string `json:"avatar_url" binding:"max=255"`
	Phone     string `json:"phone" binding:"max=32"`
	Email     string `json:"email" binding:"omitempty,email,max=254"`
	Age       int    `json:"age" binding:"gte=0,lte=150"`
	Gender    uint8  `json:"gender" binding:"oneof=0 1 2"`
	Username  string `json:"username" binding:"required,min=3,max=32"`
	Nickname  string `json:"nickname" binding:"max=50"`
	Password  string `json:"password" binding:"required,min=8,max=72"`
}
type UpdateReq struct {
	Nickname  *string `json:"nickname" binding:"omitempty,max=50"`
	AvatarURL *string `json:"avatar_url" binding:"omitempty,max=255"`
	Phone     *string `json:"phone" binding:"omitempty,max=32"`
	Email     *string `json:"email" binding:"omitempty,max=254"`
	Age       *int    `json:"age" binding:"omitempty,gte=0,lte=150"`
	Gender    *int    `json:"gender" binding:"omitempty,oneof=0 1 2"`
}
type StatusReq struct {
	Status *int `json:"status" binding:"required,oneof=0 1"`
}

type PasswordReq struct {
	Password             string `json:"password" binding:"required,min=8,max=72"`
	PasswordConfirmation string `json:"password_confirmation" binding:"required,eqfield=Password"`
}
