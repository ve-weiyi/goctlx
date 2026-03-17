package ts

import (
	"strings"
	"unicode"
)

// ApiExport 表示一个 API 导出信息
type ApiExport struct {
	FileName string // 文件名（不含扩展名），如 "payment_package"
	ApiName  string // API 常量名，如 "PaymentPackageAPI"
}

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

// ConvertPathToPascalCase 将路径转换为 PascalCase 标识符
// 例如: "Payment/package_" -> "PaymentPackage"
//
//	"user/profile" -> "UserProfile"
func ConvertPathToPascalCase(path string) string {
	// 将斜杠和下划线都替换为空格，作为分隔符
	path = strings.ReplaceAll(path, "/", " ")
	path = strings.ReplaceAll(path, "_", " ")

	// 分割成单词
	words := strings.Fields(path)

	// 将每个单词首字母大写
	var result strings.Builder
	for _, word := range words {
		if word == "" {
			continue
		}
		// 将首字母大写，其余字母小写
		runes := []rune(word)
		runes[0] = unicode.ToUpper(runes[0])
		for i := 1; i < len(runes); i++ {
			runes[i] = unicode.ToLower(runes[i])
		}
		result.WriteString(string(runes))
	}

	return result.String()
}

// ConvertPathToSnakeCase 将路径转换为 snake_case 文件名
// 例如: "Payment/package_" -> "payment_package"
//
//	"User/Profile" -> "user_profile"
func ConvertPathToSnakeCase(path string) string {
	// 将斜杠替换为下划线
	path = strings.ReplaceAll(path, "/", "_")

	// 转换为小写
	path = strings.ToLower(path)

	// 移除连续的下划线
	for strings.Contains(path, "__") {
		path = strings.ReplaceAll(path, "__", "_")
	}

	// 移除首尾的下划线
	path = strings.Trim(path, "_")

	return path
}
