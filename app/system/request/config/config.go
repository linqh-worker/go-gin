package config

type ItemReq struct {
	Key     string `json:"key" binding:"required,max=128"`
	Value   string `json:"value" binding:"max=16000"`
	Version int64  `json:"version" binding:"required,gte=1"`
}
type SaveReq struct {
	Items []ItemReq `json:"items" binding:"required,min=1,max=100,dive"`
}

type GroupReq struct {
	Code    string `json:"code" binding:"max=64"`
	Name    string `json:"name" binding:"required,max=100"`
	Version int64  `json:"version" binding:"gte=0"`
}
type VersionReq struct {
	Version int64 `json:"version" binding:"required,gte=1"`
}
type DefinitionReq struct {
	GroupCode string `json:"group_code" binding:"required,max=64"`
	Key       string `json:"key" binding:"required,max=128"`
	Name      string `json:"name" binding:"required,max=100"`
	Value     string `json:"value" binding:"max=16000"`
	ValueType string `json:"value_type" binding:"required,oneof=string integer boolean json"`
	InputType string `json:"input_type" binding:"required,oneof=input textarea number switch json"`
	IsPublic  bool   `json:"is_public"`
	IsBuiltin bool   `json:"is_builtin"`
	Sort      int    `json:"sort" binding:"gte=0,lte=1000000"`
	Remark    string `json:"remark" binding:"max=500"`
	Version   int64  `json:"version" binding:"gte=0"`
}
