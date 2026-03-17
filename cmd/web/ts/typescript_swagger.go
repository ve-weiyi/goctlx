package ts

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	apispec2 "github.com/ve-weiyi/goctlx/parserx/apispec"
)

var typescriptSwaggerFlags = struct {
	ApiFile string
	OutPath string
}{
	ApiFile: "test.api",
	OutPath: "./",
}

func NewTypescriptSwaggerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "swagger",
		Short: "根据 swagger.json 生成 TypeScript 代码",
		RunE:  runTypescriptSwagger,
	}

	cmd.Flags().StringVarP(&typescriptSwaggerFlags.ApiFile, "api-file", "f", typescriptSwaggerFlags.ApiFile, "Swagger文件路径")
	cmd.Flags().StringVarP(&typescriptSwaggerFlags.OutPath, "out-path", "o", typescriptSwaggerFlags.OutPath, "输出目录路径")

	return cmd
}

func runTypescriptSwagger(cmd *cobra.Command, args []string) error {
	fmt.Println("===== 命令参数 =====")
	fmt.Printf("api-file: %s\n", typescriptSwaggerFlags.ApiFile)
	fmt.Printf("out-path: %s\n", typescriptSwaggerFlags.OutPath)
	fmt.Println("====================")

	// 解析 swagger.json
	apiData, err := apispec2.ParseSwaggerFromFile(typescriptSwaggerFlags.ApiFile)
	if err != nil {
		return err
	}

	// 创建输出目录
	if err := os.MkdirAll(typescriptSwaggerFlags.OutPath, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// 生成类型定义文件
	typesFile := filepath.Join(typescriptSwaggerFlags.OutPath, "types.ts")
	if err := generateTypesFile(typesFile, apiData); err != nil {
		return fmt.Errorf("failed to generate types file: %w", err)
	}
	fmt.Printf("✅ Generated: %s\n", typesFile)

	// 收集所有 API 导出信息
	var apiExports []ApiExport

	// 为每个分组生成独立的 API 文件
	for _, group := range apiData.ApiGroups {
		fileName := fmt.Sprintf("%s.ts", strings.ToLower(group.Prefix))
		outputFile := filepath.Join(typescriptSwaggerFlags.OutPath, fileName)
		if err := generateApiFile(outputFile, group); err != nil {
			return fmt.Errorf("failed to generate api file: %w", err)
		}
		fmt.Printf("✅ Generated: %s\n", outputFile)

		// 收集导出信息
		apiName := ConvertPathToPascalCase(group.Name) + "API"
		apiExports = append(apiExports, ApiExport{
			FileName: strings.TrimSuffix(fileName, ".ts"),
			ApiName:  apiName,
		})
	}

	// 生成 index.ts 文件
	indexFile := filepath.Join(typescriptSwaggerFlags.OutPath, "index.ts")
	if err := generateIndexFile(indexFile, apiExports); err != nil {
		return fmt.Errorf("failed to generate index file: %w", err)
	}
	fmt.Printf("✅ Generated: %s\n", indexFile)

	return nil
}

// ============ 代码生成 ============

func generateTypesFile(filePath string, data *apispec2.ApiService) error {
	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	for _, t := range data.Types {
		if t.Comment != "" {
			comment := strings.TrimSpace(strings.TrimPrefix(t.Comment, "//"))
			fmt.Fprintf(f, "// %s\n", comment)
		}

		if len(t.Extends) > 0 {
			extends := strings.Join(t.Extends, ", ")
			fmt.Fprintf(f, "export interface %s extends %s {\n", t.Name, extends)
		} else {
			fmt.Fprintf(f, "export interface %s {\n", t.Name)
		}

		for _, field := range t.Fields {
			if field.Name == "" {
				continue
			}
			nullable := ""
			if field.Nullable {
				nullable = "?"
			}
			comment := ""
			if field.Comment != "" {
				comment = " // " + strings.TrimSpace(strings.TrimPrefix(field.Comment, "//"))
			}
			tsType := ConvertGoTypeToTsType(field.Type)
			fmt.Fprintf(f, "  %s%s: %s;%s\n", field.Name, nullable, tsType, comment)
		}
		fmt.Fprintln(f, "}")
		fmt.Fprintln(f)
	}
	return nil
}

func generateApiFile(filePath string, group apispec2.ApiGroup) error {
	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	// Import
	fmt.Fprintln(f, `import request from "@/utils/request";`)

	// 收集所有使用的类型
	typeSet := make(map[string]bool)
	for _, route := range group.Routes {
		if route.Request != "" {
			typeSet[extractBaseTypeName(route.Request)] = true
		}
		if route.Response != "" {
			typeSet[extractBaseTypeName(route.Response)] = true
		}
	}

	// 生成类型导入
	if len(typeSet) > 0 {
		// 按字母排序
		typeNames := make([]string, 0, len(typeSet))
		for typeName := range typeSet {
			typeNames = append(typeNames, typeName)
		}
		sort.Strings(typeNames)

		fmt.Fprintln(f, `import type {`)
		for _, typeName := range typeNames {
			fmt.Fprintf(f, "  %s,\n", typeName)
		}
		fmt.Fprintln(f, `} from "./types";`)
	}
	fmt.Fprintln(f)

	// API 对象
	// 使用辅助函数将路径转换为 PascalCase 标识符
	// 例如: "Payment/package_" -> "PaymentPackage"
	groupName := ConvertPathToPascalCase(group.Name)
	if groupName == "" {
		groupName = "Default"
	}

	if group.Tag != "" {
		fmt.Fprintf(f, "/** %s */\n", group.Tag)
	}
	fmt.Fprintf(f, "export const %sAPI = {\n", groupName)

	for _, route := range group.Routes {
		generateApiMethod(f, route, group)
	}

	fmt.Fprintln(f, "};")
	return nil
}

func generateApiMethod(f *os.File, route apispec2.Route, group apispec2.ApiGroup) {
	if route.Summary != "" {
		summary := strings.Trim(route.Summary, "\"")
		fmt.Fprintf(f, "  /** %s */\n", summary)
	}

	isGetRequest := strings.ToUpper(route.Method) == "GET"

	// 函数参数
	var params []string
	if route.Request != "" {
		if isGetRequest {
			params = append(params, fmt.Sprintf("params?: %s", route.Request))
		} else {
			params = append(params, fmt.Sprintf("data?: %s", route.Request))
		}
	}
	reqParam := strings.Join(params, ", ")

	response := "any"
	if route.Response != "" {
		response = ConvertGoTypeToTsType(route.Response)
	}

	fmt.Fprintf(f, "  %s(%s): Promise<IApiResponse<%s>> {\n", route.Handler, reqParam, response)

	// Request 调用
	fmt.Fprintln(f, "    return request({")
	// 拼接 prefix 和 path
	url := route.Path
	if group.Prefix != "" && group.Prefix != "default" {
		url = group.Prefix + route.Path
	}
	// 转换路径参数 :id -> ${data.id}
	if strings.Contains(url, ":") {
		// 使用模板字符串，从 data 中提取参数
		paramSource := "data"
		if isGetRequest {
			paramSource = "params"
		}
		url = strings.ReplaceAll(url, ":", "${"+paramSource+".")
		// 为每个参数添加结束符
		parts := strings.Split(url, "${")
		for i := 1; i < len(parts); i++ {
			idx := strings.IndexAny(parts[i], "/")
			if idx == -1 {
				parts[i] = parts[i] + "}"
			} else {
				parts[i] = parts[i][:idx] + "}" + parts[i][idx:]
			}
		}
		url = strings.Join(parts, "${")
	}
	// 统一使用模板字符串
	fmt.Fprintf(f, "      url: `%s`,\n", url)
	fmt.Fprintf(f, "      method: \"%s\",\n", strings.ToUpper(route.Method))

	if route.Request != "" {
		if isGetRequest {
			fmt.Fprintln(f, "      params: params,")
		} else {
			fmt.Fprintln(f, "      data: data,")
		}
	}

	fmt.Fprintln(f, "    });")
	fmt.Fprintln(f, "  },")
	fmt.Fprintln(f)
}

func extractBaseTypeName(typeName string) string {
	// 移除 []*  []  * 等前缀
	typeName = strings.TrimPrefix(typeName, "[]*")
	typeName = strings.TrimPrefix(typeName, "[]")
	typeName = strings.TrimPrefix(typeName, "*")
	return typeName
}

func extractPathParams(path string) []string {
	var params []string
	parts := strings.Split(path, "/")
	for _, part := range parts {
		if strings.HasPrefix(part, ":") {
			params = append(params, strings.TrimPrefix(part, ":"))
		}
	}
	return params
}

// generateIndexFile 生成 index.ts 文件，统一导出所有 API
func generateIndexFile(filePath string, apiExports []ApiExport) error {
	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	// 导出 types
	fmt.Fprintln(f, "export * from './types';")
	fmt.Fprintln(f)

	// 导出所有 API
	sort.Slice(apiExports, func(i, j int) bool {
		return apiExports[i].ApiName < apiExports[j].ApiName
	})
	for _, export := range apiExports {
		fmt.Fprintf(f, "export { %s } from './%s';\n", export.ApiName, export.FileName)
	}

	return nil
}
