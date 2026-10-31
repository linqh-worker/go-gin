package menu

type SaveReq struct {
	ParentID       *int   `json:"parent_id" binding:"required,gte=0"`
	Code           string `json:"code" binding:"required,max=64"`
	Name           string `json:"name" binding:"required,max=100"`
	Type           string `json:"type" binding:"required,oneof=directory menu button"`
	Path           string `json:"path" binding:"max=255"`
	Component      string `json:"component" binding:"max=255"`
	Icon           string `json:"icon" binding:"max=100"`
	PermissionCode string `json:"permission_code" binding:"max=100"`
	Sort           int    `json:"sort" binding:"gte=0,lte=1000000"`
	Status         *int   `json:"status" binding:"required,oneof=0 1"`
}
type RoleMenusReq struct {
	MenuIDs []int `json:"menu_ids" binding:"required,max=1000"`
}
