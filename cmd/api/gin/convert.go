package gin

// Unused - 这些类型用于旧的 API 解析方式，保留以供参考
// 新实现请使用 gen_apiservice.go 中的函数

type GroupRoute struct {
	Name       string
	Prefix     string
	Middleware []string
	Routes     []Route
}

type Route struct {
	Doc      string
	Handler  string
	Path     string
	Method   string
	Request  string
	Response string
}

type GroupType map[string]map[string]interface{}
