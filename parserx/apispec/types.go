package apispec

// ParamLocation 表示参数来源位置
type ParamLocation string

const (
	LocationBody   ParamLocation = "body"   // json tag → JSON 请求体
	LocationPath   ParamLocation = "path"   // path tag → URL 路径参数
	LocationForm   ParamLocation = "form"   // form tag → 表单字段
	LocationHeader ParamLocation = "header" // header tag → 请求头
)

// ApiService 表示整个 API 服务
type ApiService struct {
	Name   string
	Types  []Type
	Groups []ApiGroup
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
	Name     string        // 序列化名称（从 tag 中提取）
	Type     string        // Go 类型名
	Location ParamLocation // 参数来源：body / path / form / header
	Optional bool
	Comment  string
}

// ApiGroup 表示 API 分组
type ApiGroup struct {
	Name       string // 分组名，如 "blog/article"
	Label      string // 分组标签，如 "文章"
	Prefix     string // URL 前缀，如 "/api/v1"
	Middleware []string
	Routes     []Route
}

// Route 表示一个 API 路由
type Route struct {
	Method   string  // HTTP 方法
	Path     string  // 路由路径
	Handler  string  // 方法名（camelCase）
	Summary  string  // 接口说明
	Request  string  // 请求类型名（空表示无请求体）
	Response string  // 响应类型名（空表示 any）
	Params   []Field // 请求参数（从 Request 类型的 Fields 预计算）
}
