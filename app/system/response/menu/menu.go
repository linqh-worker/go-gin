package menu

type Node struct {
	ID             int    `json:"id"`
	ParentID       int    `json:"parent_id"`
	Code           string `json:"code"`
	Name           string `json:"name"`
	Type           string `json:"type"`
	Path           string `json:"path"`
	Component      string `json:"component"`
	Icon           string `json:"icon"`
	PermissionCode string `json:"permission_code"`
	Sort           int    `json:"sort"`
	Status         int    `json:"status"`
	Children       []Node `json:"children"`
}
