package rbac

type RoleReq struct {
	Code          string `json:"code" binding:"max=64"`
	Name          string `json:"name" binding:"required,max=100"`
	Status        *int   `json:"status" binding:"required,oneof=0 1"`
	PermissionIDs []int  `json:"permission_ids" binding:"required,max=100"`
}
type AdminRolesReq struct {
	RoleIDs []int `json:"role_ids" binding:"required,max=100"`
}
