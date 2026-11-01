package menu

import (
	"context"
	"go-gin/app/system/dao"
	request "go-gin/app/system/request/menu"
	response "go-gin/app/system/response/menu"
	"go-gin/internal/errorx"
	"regexp"
	"strings"
)

func Save(ctx context.Context, id int, r request.SaveReq) error {
	if r.ParentID == nil || r.Status == nil {
		return errorx.New(10001, "parent_id和status不能为空")
	}
	r.Code = strings.TrimSpace(r.Code)
	r.Name = strings.TrimSpace(r.Name)
	r.Path = strings.TrimSpace(r.Path)
	if !regexp.MustCompile(`^[a-z][a-z0-9_.:-]{2,63}$`).MatchString(r.Code) || r.Name == "" {
		return errorx.New(10001, "菜单编码须为3至64位字母数字及_.:-，以小写字母开头；名称不能为空")
	}
	if r.Type == "menu" && (!strings.HasPrefix(r.Path, "/") || strings.HasPrefix(r.Path, "//")) {
		return errorx.New(10001, "菜单路径必须以单个 / 开头")
	}
	if (r.Type == "button" || r.Type == "api") && (r.Method == "" || r.APIPath == "" || r.Path != "" || r.Component != "") {
		return errorx.New(10001, "按钮或接口节点必须配置method和api_path，且不能填写前端路径和组件")
	}
	if r.Type != "directory" && r.Type != "menu" && r.Type != "button" && r.Type != "api" {
		return errorx.New(10001, "菜单类型不正确")
	}
	if r.Method != "" || r.APIPath != "" {
		if !regexp.MustCompile(`^/system/[A-Za-z0-9_/:.\-]+$`).MatchString(r.APIPath) || r.Method == "" || strings.Contains(r.APIPath, "//") {
			return errorx.New(10001, "method和api_path必须成对填写，接口使用完整后台路由模板")
		}
	}
	if r.Type == "directory" && (r.Method != "" || r.APIPath != "") {
		return errorx.New(10001, "目录不能绑定接口")
	}
	m := dao.Menu{ID: id, ParentID: *r.ParentID, Code: r.Code, Name: r.Name, Type: r.Type, Path: r.Path, Component: r.Component, Icon: r.Icon, Method: r.Method, APIPath: r.APIPath, Sort: r.Sort, Status: *r.Status}
	return dao.SaveMenu(ctx, m, validateTree)
}

func validateTree(rows []dao.Menu, m dao.Menu) error {
	nodes := map[int]dao.Menu{}
	found := m.ID == 0
	for _, r := range rows {
		if r.Code == m.Code && r.ID != m.ID {
			return errorx.New(10001, "菜单编码已存在")
		}
		nodes[r.ID] = r
		if r.ID == m.ID {
			found = true
		}
	}
	if !found {
		return errorx.New(10001, "菜单不存在")
	}
	if m.ID == 0 {
		if len(rows) >= 5000 {
			return errorx.New(10001, "菜单节点不能超过5000个")
		}
		m.ID = -1
	}
	nodes[m.ID] = m
	for _, n := range nodes {
		if n.ParentID == 0 {
			if n.Type == "button" || n.Type == "api" {
				return errorx.New(10001, "按钮必须挂在菜单下面")
			}
		} else {
			parent, ok := nodes[n.ParentID]
			if !ok {
				return errorx.New(10001, "父节点不存在")
			}
			if parent.Type == "button" || parent.Type == "api" || parent.Type == "menu" && n.Type != "button" && n.Type != "api" || parent.Type == "directory" && (n.Type == "button" || n.Type == "api") {
				return errorx.New(10001, "目录下放目录或菜单，菜单下放按钮，按钮不能有子节点")
			}
		}
		seen := map[int]bool{}
		current := n.ID
		depth := 0
		for current != 0 {
			if seen[current] {
				return errorx.New(10001, "不能把节点挂到自己或自己的子节点下")
			}
			seen[current] = true
			depth++
			if depth > 10 {
				return errorx.New(10001, "菜单层级最多10层")
			}
			row, ok := nodes[current]
			if !ok {
				return errorx.New(10001, "父节点不存在")
			}
			current = row.ParentID
		}
	}
	return nil
}

func Tree(ctx context.Context) ([]response.Node, error) {
	rows, err := dao.Menus(ctx)
	if err != nil {
		return nil, err
	}
	return buildTree(rows, nil), nil
}
func CurrentTree(ctx context.Context, adminID int) ([]response.Node, error) {
	access, err := dao.LoadAccess(ctx, adminID)
	if err != nil {
		return nil, err
	}
	rows, err := dao.Menus(ctx)
	if err != nil {
		return nil, err
	}
	selected := map[int]bool{}
	super := false
	for _, r := range access.Roles {
		if r == "super_admin" {
			super = true
		}
	}
	if super {
		for _, r := range rows {
			selected[r.ID] = true
		}
	} else {
		ids, err := dao.GrantedMenus(ctx, adminID)
		if err != nil {
			return nil, err
		}
		for _, id := range ids {
			selected[id] = true
		}
	}
	// 仅补齐已授权节点的启用祖先，不扩展到兄弟节点或整棵子树。
	byID := map[int]dao.Menu{}
	for _, r := range rows {
		byID[r.ID] = r
	}
	allowed := map[int]bool{}
	for id := range selected {
		chain := []int{}
		seen := map[int]bool{}
		current := id
		valid := true
		for current != 0 {
			n, ok := byID[current]
			if !ok || n.Status != 1 || seen[current] || len(chain) >= 10 {
				valid = false
				break
			}
			seen[current] = true
			chain = append(chain, current)
			current = n.ParentID
		}
		if valid {
			for _, n := range chain {
				allowed[n] = true
			}
		}
	}
	return buildTree(rows, allowed), nil
}

func buildTree(rows []dao.Menu, allowed map[int]bool) []response.Node {
	children := map[int][]dao.Menu{}
	for _, r := range rows {
		if allowed == nil || allowed[r.ID] {
			children[r.ParentID] = append(children[r.ParentID], r)
		}
	}
	var walk func(int, int) []response.Node
	walk = func(parent, depth int) []response.Node {
		out := []response.Node{}
		if depth > 10 {
			return out
		}
		for _, r := range children[parent] {
			out = append(out, response.Node{ID: r.ID, ParentID: r.ParentID, Code: r.Code, Name: r.Name, Type: r.Type, Path: r.Path, Component: r.Component, Icon: r.Icon, Method: r.Method, APIPath: r.APIPath, Sort: r.Sort, Status: r.Status, Children: walk(r.ID, depth+1)})
		}
		return out
	}
	return walk(0, 1)
}
func Delete(ctx context.Context, id int) error             { return dao.DeleteMenu(ctx, id) }
func RoleMenus(ctx context.Context, id int) ([]int, error) { return dao.RoleMenus(ctx, id) }
func SetRoleMenus(ctx context.Context, id int, ids []int) error {
	return dao.SetRoleMenus(ctx, id, ids)
}

// CanAccess 只认角色直接关联的接口节点；显示树补齐的祖先不等于接口授权。
func CanAccess(ctx context.Context, method, path string, menuIDs []int, super bool) (bool, error) {
	rows, err := dao.Menus(ctx)
	if err != nil {
		return false, err
	}
	byID := map[int]dao.Menu{}
	granted := map[int]bool{}
	for _, id := range menuIDs {
		granted[id] = true
	}
	for _, row := range rows {
		byID[row.ID] = row
	}
	for _, row := range rows {
		if row.Method != method || row.APIPath != path || (!super && !granted[row.ID]) {
			continue
		}
		current := row.ID
		seen := map[int]bool{}
		valid := true
		for current != 0 {
			node, ok := byID[current]
			if !ok || node.Status != 1 || seen[current] || len(seen) >= 10 {
				valid = false
				break
			}
			seen[current] = true
			current = node.ParentID
		}
		if valid {
			return true, nil
		}
	}
	return false, nil
}
