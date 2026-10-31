# 后台菜单树

菜单通过 `parent_id` 表示上下级：根节点为0；接口通过 `children` 返回树，叶子节点返回空数组。

```text
系统管理（directory）
├─ 管理员管理（menu）
│  └─ 查看管理员（button）
└─ 角色管理（menu）
   └─ 查看角色（button）
```

目录下面可以放目录或菜单；菜单下面只放按钮；按钮不允许有子节点。最多10层、5000个节点。同级按 sort 升序、再按 id 升序排列。

## 数据表与初始化

`system_menu` 保存树节点；`system_role_menu` 保存角色与节点关联。菜单与接口权限使用不同的表、不同的授权接口。

```powershell
go run ./cmd/migrate -f ./.env
```

初始化只创建表，不预填不存在的前端页面；尚未创建节点或尚未分配菜单时，个人菜单返回 `[]`。

## 接口

均需要后台 Bearer Token。除个人菜单外，管理操作仅超级管理员可用。

| 方法 | 路径 | 用途 |
| --- | --- | --- |
| GET | /system/menus/mine | 当前管理员可见树，自动补齐父节点 |
| GET | /system/menus | 超管查看完整树，包含停用节点 |
| POST | /system/menus | 新建节点 |
| PUT | /system/menus/:id | 完整修改节点及父级 |
| DELETE | /system/menus/:id | 删除叶子节点，并清理角色关联 |
| GET | /system/roles/:id/menus | 查看角色直接授予的节点ID |
| PUT | /system/roles/:id/menus | 完整替换角色菜单授权 |

## 创建节点

创建根目录：

```json
{
  "parent_id": 0,
  "code": "system_management",
  "name": "系统管理",
  "type": "directory",
  "sort": 10,
  "status": 1
}
```

请求 `GET /system/menus` 找到实际目录ID，在其下面创建菜单：

```json
{
  "parent_id": 1,
  "code": "admin_management",
  "name": "管理员管理",
  "type": "menu",
  "path": "/system/admins",
  "component": "system/admin/index",
  "icon": "user",
  "sort": 10,
  "status": 1
}
```

上面的ID、path和component是示例，ID替换为真实值，path和component对应自己的前端页面，不会自动创建页面或后端路由。

创建按钮时 parent_id 指向菜单，type 为 button，permission_code 指向已有接口权限编码，如 `system.admin.read`；按钮不能填写 path/component。编码全表唯一。

PUT 使用与 POST 相同的完整字段结构；status=0停用、1启用。改变父级时会检查父节点存在、类型符合规则、没有循环、未超过层数；修改父节点类型也不能破坏已有子节点。

## 角色授权与个人菜单

```http
PUT /system/roles/角色ID/menus
```

```json
{"menu_ids": [2, 3]}
```

ID必须替换为真实节点ID。授权是完整替换，传 `[]` 清空；重复、不存在或停用节点会被拒绝并回滚。超级管理员角色无需分配，自动拥有全部启用节点。

只选子菜单或按钮时，个人菜单接口会补齐启用的祖先，但不会授予兄弟节点。只选目录不会自动选中所有后代，需要明确传递后代ID。一个管理员多个启用角色的菜单会合并去重。

停用父节点会隐藏整棵子树，重新启用后原授权恢复；撤销角色菜单或停用角色后，下一次个人菜单请求生效。有子节点不能直接删除，需从叶子向上删除。

## 菜单与接口权限分开

`permission_code` 表示该节点关联哪个接口权限，用于前端理解按钮含义，**不会自动给角色授予接口权限**。

例如角色获得“查看管理员”按钮后，还需通过角色权限接口授予 `system.admin.read`；否则请求接口仍返回403。反过来，隐藏或删除菜单也不会撤销已经分配的接口权限；需要分别管理。

前端根据 type 渲染侧边栏或按钮，根据 `/system/me` 的 permissions 检查按钮操作权限。后端始终以 RequirePermission 校验接口，不能用菜单树代替服务端鉴权。
