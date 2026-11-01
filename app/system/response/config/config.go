package config

type Item struct {
	ID        int64   `json:"id"`
	Sort      int     `json:"sort"`
	Key       string  `json:"key"`
	Name      string  `json:"name"`
	Value     string  `json:"value"`
	ValueType string  `json:"value_type"`
	InputType string  `json:"input_type"`
	IsPublic  bool    `json:"is_public"`
	IsBuiltin bool    `json:"is_builtin"`
	Remark    string  `json:"remark"`
	Version   int64   `json:"version"`
	UpdatedAt *string `json:"updated_at"`
}
type Group struct {
	Version   int64  `json:"version"`
	IsBuiltin bool   `json:"is_builtin"`
	Code      string `json:"code"`
	Name      string `json:"name"`
	Items     []Item `json:"items"`
}
