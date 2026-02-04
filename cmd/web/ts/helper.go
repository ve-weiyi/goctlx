package ts

import "strings"

// ConvertGoTypeToTsType 将 Go 类型转换为 TypeScript 类型
func ConvertGoTypeToTsType(goType string) string {
	// 移除指针标记
	goType = strings.TrimPrefix(goType, "*")

	// 处理数组类型
	if strings.HasPrefix(goType, "[]") {
		elemType := strings.TrimPrefix(goType, "[]")
		return ConvertGoTypeToTsType(elemType) + "[]"
	}

	// 处理 map 类型
	if strings.HasPrefix(goType, "map[") {
		return "Record<string, any>"
	}

	// 基础类型映射
	switch goType {
	case "string":
		return "string"
	case "int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64":
		return "number"
	case "float32", "float64":
		return "number"
	case "bool":
		return "boolean"
	case "byte":
		return "number"
	case "interface{}", "any":
		return "any"
	default:
		// 自定义类型保持不变
		return goType
	}
}
