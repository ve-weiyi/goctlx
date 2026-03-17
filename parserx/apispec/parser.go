package apispec

import (
	"fmt"
	"strings"

	"github.com/go-openapi/loads"
	"github.com/go-openapi/spec"
)

// ParseSwaggerFromFile 从文件解析 Swagger 规范为 ApiService
func ParseSwaggerFromFile(filePath string) (*ApiService, error) {
	doc, err := loads.Spec(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load swagger spec: %w", err)
	}
	return ParseSwagger(doc.Spec()), nil
}

// ParseSwagger 解析 Swagger 规范为 ApiService
func ParseSwagger(swagger *spec.Swagger) *ApiService {
	service := &ApiService{
		Name:      "API",
		Types:     []Type{},
		ApiGroups: []ApiGroup{},
	}

	// 解析 definitions
	for name, schema := range swagger.Definitions {
		t := Type{
			Name:    name,
			Comment: schema.Description,
			Fields:  []Field{},
		}

		for propName, prop := range schema.Properties {
			field := Field{
				Name:     propName,
				Type:     SwaggerTypeToGo(prop),
				Tag:      BuildTag(propName, !Contains(schema.Required, propName)),
				Comment:  prop.Description,
				Nullable: !Contains(schema.Required, propName),
			}
			t.Fields = append(t.Fields, field)
		}
		service.Types = append(service.Types, t)
	}

	// 解析 paths
	groupMap := make(map[string]*ApiGroup)
	for path, pathItem := range swagger.Paths.Paths {
		operations := map[string]*spec.Operation{
			"GET":    pathItem.Get,
			"POST":   pathItem.Post,
			"PUT":    pathItem.Put,
			"DELETE": pathItem.Delete,
			"PATCH":  pathItem.Patch,
		}

		for method, op := range operations {
			if op == nil {
				continue
			}

			tag := "default"
			if len(op.Tags) > 0 {
				tag = op.Tags[0]
			}

			groupKey := ExtractGroupKey(op.ID, tag)

			if groupMap[groupKey] == nil {
				groupMap[groupKey] = &ApiGroup{
					Prefix:     groupKey,
					Tag:        tag,
					Middleware: []string{},
					Routes:     []Route{},
				}
			}

			route := ParseOperation(op, path, method)
			groupMap[groupKey].Routes = append(groupMap[groupKey].Routes, route)
		}
	}

	for _, group := range groupMap {
		service.ApiGroups = append(service.ApiGroups, *group)
	}

	return service
}

// ExtractGroupKey 从 operationID 提取分组键
func ExtractGroupKey(operationID, defaultTag string) string {
	if operationID == "" {
		return defaultTag
	}

	parts := strings.Split(operationID, "_")
	if len(parts) > 1 {
		return parts[0]
	}

	for i, r := range operationID {
		if i > 0 && r >= 'A' && r <= 'Z' {
			return strings.ToLower(operationID[:i])
		}
	}

	return defaultTag
}

// ExtractHandlerName 从 operationID 提取方法名（移除分组前缀，首字母小写）
func ExtractHandlerName(operationID string) string {
	if operationID == "" {
		return operationID
	}

	// 按下划线分割
	parts := strings.Split(operationID, "_")
	if len(parts) > 1 {
		handlerName := strings.Join(parts[1:], "_")
		return lowerFirst(handlerName)
	}

	// 按驼峰分割
	for i, r := range operationID {
		if i > 0 && r >= 'A' && r <= 'Z' {
			return lowerFirst(operationID[i:])
		}
	}

	return lowerFirst(operationID)
}

// lowerFirst 将首字母转为小写
func lowerFirst(s string) string {
	if s == "" {
		return s
	}
	return strings.ToLower(s[:1]) + s[1:]
}

// ParseOperation 解析单个操作
func ParseOperation(op *spec.Operation, path, method string) Route {
	handlerName := ExtractHandlerName(op.ID)
	if handlerName == "" {
		handlerName = op.ID
	}

	route := Route{
		Handler:     handlerName,
		Summary:     op.Summary,
		Path:        path,
		Method:      strings.ToUpper(method),
		QueryParams: []QueryParam{},
	}

	// 解析请求参数
	for _, param := range op.Parameters {
		if param.In == "body" && param.Schema != nil && param.Schema.Ref.String() != "" {
			route.Request = GetRefName(param.Schema.Ref.String())
			break
		} else if param.In == "query" {
			qp := QueryParam{
				Name:        param.Name,
				Type:        param.Type,
				Description: param.Description,
				Required:    param.Required,
			}
			route.QueryParams = append(route.QueryParams, qp)
		}
	}

	// 解析响应
	if resp, ok := op.Responses.StatusCodeResponses[200]; ok {
		if resp.Schema != nil {
			if resp.Schema.Ref.String() != "" {
				route.Response = GetRefName(resp.Schema.Ref.String())
			} else if resp.Schema.Properties != nil {
				if dataProp, ok := resp.Schema.Properties["data"]; ok {
					if dataProp.Ref.String() != "" {
						route.Response = GetRefName(dataProp.Ref.String())
					}
				}
			}
		}
	}

	return route
}

// SwaggerTypeToGo 将 Swagger 类型转换为 Go 类型
func SwaggerTypeToGo(schema spec.Schema) string {
	if schema.Ref.String() != "" {
		return GetRefName(schema.Ref.String())
	}

	if len(schema.Type) == 0 {
		return "interface{}"
	}

	switch schema.Type[0] {
	case "string":
		return "string"
	case "integer":
		if schema.Format == "int64" {
			return "int64"
		}
		return "int"
	case "number":
		if schema.Format == "float" {
			return "float32"
		}
		return "float64"
	case "boolean":
		return "bool"
	case "array":
		if schema.Items != nil && schema.Items.Schema != nil {
			return "[]" + SwaggerTypeToGo(*schema.Items.Schema)
		}
		return "[]interface{}"
	case "object":
		return "map[string]interface{}"
	default:
		return "interface{}"
	}
}

// BuildTag 构建字段标签
func BuildTag(fieldName string, optional bool) string {
	jsonTag := fieldName
	if optional {
		jsonTag += ",omitempty"
	}
	return fmt.Sprintf("`json:\"%s\"`", jsonTag)
}

// GetRefName 从引用路径中提取名称
func GetRefName(ref string) string {
	parts := strings.Split(ref, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return ref
}

// Contains 检查切片是否包含指定元素
func Contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
