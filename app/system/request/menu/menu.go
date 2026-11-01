package menu

type SaveReq struct {
	ParentID  *int   `json:"parent_id" binding:"required,gte=0"`
	Code      string `json:"code" binding:"required,max=64"`
	Name      string `json:"name" binding:"required,max=100"`
	Type      string `json:"type" binding:"required,oneof=directory menu button api"`
	Path      string `json:"path" binding:"max=255"`
	Component string `json:"component" binding:"max=255"`
	Icon      string `json:"icon" binding:"max=100"`
	Method    string `json:"method" binding:"omitempty,oneof=GET POST PUT PATCH DELETE HEAD OPTIONS"`
	APIPath   string `json:"api_path" binding:"max=255"`
	Sort      int    `json:"sort" binding:"gte=0,lte=1000000"`
	Status    *int   `json:"status" binding:"required,oneof=0 1"`
}
type RoleMenusReq struct {
	MenuIDs []int `json:"menu_ids" binding:"required,max=1000"`
}
