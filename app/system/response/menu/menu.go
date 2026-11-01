package menu

type Node struct {
	ID        int    `json:"id"`
	ParentID  int    `json:"parent_id"`
	Code      string `json:"code"`
	Name      string `json:"name"`
	Type      string `json:"type"`
	Path      string `json:"path"`
	Component string `json:"component"`
	Icon      string `json:"icon"`
	Method    string `json:"method" binding:"omitempty,oneof=GET POST PUT PATCH DELETE HEAD OPTIONS"`
	APIPath   string `json:"api_path" binding:"max=255"`
	Sort      int    `json:"sort"`
	Status    int    `json:"status"`
	Children  []Node `json:"children"`
}
