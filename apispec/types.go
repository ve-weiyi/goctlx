package apispec

// ApiService 表示整个 API 服务
type ApiService struct {
	Name      string
	Types     []Type
	ApiGroups []ApiGroup
}

// Type 表示一个类型定义
type Type struct {
	Name    string
	Comment string
	Extends []string
	Fields  []Field
}

// Field 表示类型的字段
type Field struct {
	Name     string
	Type     string
	Tag      string
	Comment  string
	Nullable bool
}

// ApiGroup 表示 API 分组
type ApiGroup struct {
	Name       string
	Prefix     string
	Tag        string
	Middleware []string
	Routes     []Route
}

// Route 表示一个 API 路由
type Route struct {
	Handler     string
	Summary     string
	Path        string
	Method      string
	Request     string
	Response    string
	QueryParams []QueryParam
}

// QueryParam 表示查询参数
type QueryParam struct {
	Name        string
	Type        string
	Description string
	Required    bool
}
