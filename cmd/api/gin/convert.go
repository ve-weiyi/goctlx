package gin

import (
	"sort"
	"strings"

	"github.com/zeromicro/go-zero/tools/goctl/util/stringx"

	"github.com/ve-weiyi/goctlx/parserx/apispec"
)

type GinGroupRoute struct {
	Name       string
	Prefix     string
	Middleware []string
	Routes     []GinRoute
}

type GinRoute struct {
	Doc      string
	Handler  string
	Path     string
	Method   string
	Request  string
	Response string
}

func ConvertRouteGroups(service *apispec.ApiService) map[string][]GinGroupRoute {
	out := make(map[string][]GinGroupRoute)
	for _, group := range service.Groups {
		name := group.Name
		if idx := strings.LastIndex(name, "/"); idx != -1 {
			name = name[idx+1:]
		}
		name = stringx.From(name).ToCamel()
		if name == "" {
			name = service.Name
		}

		var routes []GinRoute
		for _, r := range group.Routes {
			handler := strings.ToUpper(r.Handler[:1]) + r.Handler[1:]
			doc := strings.Trim(r.Summary, "\"")
			rt := GinRoute{
				Method:   strings.ToUpper(r.Method),
				Path:     r.Path,
				Handler:  handler,
				Doc:      doc,
				Request:  r.Request,
				Response: r.Response,
			}
			routes = append(routes, rt)
		}

		gr := GinGroupRoute{
			Name:       name,
			Prefix:     group.Prefix,
			Middleware: group.Middleware,
			Routes:     routes,
		}
		out[name] = append(out[name], gr)
	}
	return out
}

// GroupTypes 将类型按路由引用关系分配到分组
func GroupTypes(service *apispec.ApiService) map[string]map[string]apispec.Type {
	defaultGroup := "types"

	// 收集每个类型被哪些分组引用
	typeToGroups := make(map[string]map[string]bool)
	for _, group := range service.Groups {
		groupName := group.Name
		if groupName == "" {
			groupName = defaultGroup
		}

		for _, route := range group.Routes {
			if route.Request != "" {
				addTypeToGroup(typeToGroups, route.Request, groupName)
			}
			if route.Response != "" {
				addTypeToGroup(typeToGroups, route.Response, groupName)
			}
		}
	}

	// 将类型分配到对应分组
	groupTypes := make(map[string]map[string]apispec.Type)
	groupTypes[defaultGroup] = make(map[string]apispec.Type)

	for _, typ := range service.Types {
		typeName := typ.Name
		groups := typeToGroups[typeName]

		var targetGroup string
		if len(groups) == 1 {
			for g := range groups {
				targetGroup = g
				break
			}
		} else {
			targetGroup = defaultGroup
		}

		if _, ok := groupTypes[targetGroup]; !ok {
			groupTypes[targetGroup] = make(map[string]apispec.Type)
		}
		groupTypes[targetGroup][typeName] = typ
	}

	return groupTypes
}

func addTypeToGroup(m map[string]map[string]bool, typeName, groupName string) {
	// 去掉指针和数组前缀
	typeName = strings.TrimPrefix(typeName, "[]*")
	typeName = strings.TrimPrefix(typeName, "[]")
	typeName = strings.TrimPrefix(typeName, "*")
	if typeName == "" {
		return
	}
	if _, ok := m[typeName]; !ok {
		m[typeName] = make(map[string]bool)
	}
	m[typeName][groupName] = true
}

// buildTypes 根据 apispec.Type 生成 Go struct 定义字符串
func buildTypes(tp apispec.Type) string {
	var b strings.Builder

	// 注释
	if tp.Comment != "" {
		comment := strings.TrimPrefix(tp.Comment, "//")
		b.WriteString("// " + strings.TrimSpace(comment) + "\n")
	}

	b.WriteString("type " + tp.Name + " struct {\n")

	// 内嵌类型
	for _, ext := range tp.Extends {
		b.WriteString("\t" + ext + "\n")
	}

	// 字段
	for _, f := range tp.Fields {
		if f.Name == "" {
			continue
		}
		fieldName := stringx.From(f.Name).ToCamel()
		goType := f.Type
		if f.Optional && !strings.HasPrefix(goType, "*") {
			goType = "*" + goType
		}

		// 构建 tag
		tag := buildFieldTag(f)

		b.WriteString("\t" + fieldName + " " + goType + " " + tag)

		if f.Comment != "" {
			comment := strings.TrimPrefix(f.Comment, "//")
			b.WriteString(" // " + strings.TrimSpace(comment))
		}
		b.WriteString("\n")
	}

	b.WriteString("}")
	return b.String()
}

// buildFieldTag 根据 Field 的 Location 和 Optional 构建 Go tag
func buildFieldTag(f apispec.Field) string {
	tagKey := "json"
	switch f.Location {
	case apispec.LocationForm:
		tagKey = "form"
	case apispec.LocationPath:
		tagKey = "uri"
	case apispec.LocationHeader:
		tagKey = "header"
	}

	tagValue := f.Name
	if f.Optional {
		tagValue += ",optional"
	}

	return "`" + tagKey + ":\"" + tagValue + "\"`"
}

// sortedGroups 返回按名称排序的分组名列表
func sortedGroups(groups map[string][]GinGroupRoute) []string {
	names := make([]string, 0, len(groups))
	for k := range groups {
		names = append(names, k)
	}
	sort.Strings(names)
	return names
}

// sortedTypeNames 返回按名称排序的类型名列表
func sortedTypeNames(types map[string]apispec.Type) []string {
	names := make([]string, 0, len(types))
	for k := range types {
		names = append(names, k)
	}
	sort.Strings(names)
	return names
}
