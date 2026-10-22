package ts

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ve-weiyi/goctlx/parserx/apispec"
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

	apiData, err := apispec.ParseSwaggerFromFile(typescriptSwaggerFlags.ApiFile)
	if err != nil {
		return err
	}

	return generateFromApiService(typescriptSwaggerFlags.OutPath, apiData)
}

// ============ 代码生成 ============

func generateTypesFile(filePath string, data *apispec.ApiService) error {
	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	for i, t := range data.Types {
		if i > 0 {
			fmt.Fprintln(f)
		}

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
			if field.Optional {
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
	}
	return nil
}

func generateApiFile(filePath string, group apispec.ApiGroup) error {
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
		fmt.Fprintln(f, `} from "@/api/types";`)
	}
	fmt.Fprintln(f)

	// API 对象
	// 使用辅助函数将路径转换为 PascalCase 标识符
	// 例如: "Payment/package_" -> "PaymentPackage"
	groupName := LastSegmentPascalCase(group.Name)
	if groupName == "" {
		groupName = "Default"
	}

	if group.Label != "" {
		fmt.Fprintf(f, "/** %s */\n", group.Label)
	}
	fmt.Fprintf(f, "export const %sAPI = {\n", groupName)

	for i, route := range group.Routes {
		if i > 0 {
			fmt.Fprintln(f)
		}
		generateApiMethod(f, route, group)
	}

	fmt.Fprintln(f, "};")
	return nil
}

func generateApiMethod(f *os.File, route apispec.Route, group apispec.ApiGroup) {
	writeMethodDoc(f, route)

	isGetRequest := strings.ToUpper(route.Method) == "GET"
	isFileUpload := isFileUploadRoute(route)

	// 形参名与发送方式一致：有请求体走 data，其余走 params。
	// 两者必须同名——路径参数插值、FormData 取值、request 配置三处都引用它。
	useBody := !isGetRequest && hasBodyFields(route)
	paramName := "params"
	if isFileUpload || useBody {
		paramName = "data"
	}

	fmt.Fprintf(f, "  %s(%s): Promise<ApiResponse<%s>> {\n", route.Handler, methodParams(route, paramName, isFileUpload), methodResponse(route))

	// 文件上传：构建 FormData
	if isFileUpload {
		writeFormData(f, route)
	}

	// Request 调用
	fmt.Fprintln(f, "    return request({")
	fmt.Fprintf(f, "      url: `%s`,\n", buildRouteURL(route, group, paramName))
	fmt.Fprintf(f, "      method: \"%s\",\n", strings.ToUpper(route.Method))
	writeRequestBody(f, route, paramName, isFileUpload, useBody)
	fmt.Fprintln(f, "    });")
	fmt.Fprintln(f, "  },")
}

// writeMethodDoc 输出 JSDoc 注释（契约里有 summary 时）。
func writeMethodDoc(f *os.File, route apispec.Route) {
	if route.Summary == "" {
		return
	}
	summary := strings.Trim(route.Summary, "\"")
	fmt.Fprintf(f, "  /** %s */\n", summary)
}

// methodParams 拼函数形参表：路径参数与文件上传都会无条件读取形参，
// 前者缺了会拼出 /apis/undefined，后者 data.file 会直接报错，故不标可选。
func methodParams(route apispec.Route, paramName string, isFileUpload bool) string {
	if route.Request == "" {
		return ""
	}
	optional := "?"
	if hasPathFields(route) || isFileUpload {
		optional = ""
	}
	return fmt.Sprintf("%s%s: %s", paramName, optional, route.Request)
}

// methodResponse 返回 Promise 包裹的响应类型。
func methodResponse(route apispec.Route) string {
	if route.Response == "" {
		return "any"
	}
	return ConvertGoTypeToTsType(route.Response)
}

// writeFormData 为文件上传构建 FormData；可选字段缺省时不应出现在表单里，
// 否则 undefined 会被当成字符串提交。
func writeFormData(f *os.File, route apispec.Route) {
	fmt.Fprintln(f, "    const formData = new FormData();")
	for _, param := range route.Params {
		if param.Optional {
			fmt.Fprintf(f, "    if (data.%s !== undefined) {\n", param.Name)
			fmt.Fprintf(f, "      formData.append(\"%s\", data.%s);\n", param.Name, param.Name)
			fmt.Fprintln(f, "    }")
			continue
		}
		fmt.Fprintf(f, "    formData.append(\"%s\", data.%s);\n", param.Name, param.Name)
	}
	fmt.Fprintln(f)
}

// buildRouteURL 拼接 prefix 与 path，并把路径参数 :id 换成模板插值 ${params.id}。
func buildRouteURL(route apispec.Route, group apispec.ApiGroup, paramName string) string {
	url := route.Path
	if group.Prefix != "" && group.Prefix != "default" {
		url = group.Prefix + route.Path
	}
	for _, param := range route.Params {
		if param.Location == apispec.LocationPath {
			placeholder := ":" + param.Name
			replacement := "${" + paramName + "." + param.Name + "}"
			url = strings.Replace(url, placeholder, replacement, 1)
		}
	}
	return url
}

// writeRequestBody 按契约选择 request 的传参方式。
func writeRequestBody(f *os.File, route apispec.Route, paramName string, isFileUpload, useBody bool) {
	if route.Request == "" {
		return
	}
	if isFileUpload {
		fmt.Fprintln(f, "      data: formData,")
		fmt.Fprintln(f, "      headers: {")
		fmt.Fprintln(f, "        \"Content-Type\": \"multipart/form-data\",")
		fmt.Fprintln(f, "      },")
		return
	}
	if useBody {
		fmt.Fprintln(f, "      data: data,")
		return
	}
	// 契约里没有 json 字段就没有请求体：例如 DELETE /articles/:id 的参数是
	// path:"id"，发 body 既不必要，也与「不用 DELETE 带 body」的约定相悖。
	fmt.Fprintf(f, "      params: %s,\n", paramName)
}

// hasBodyFields 判断请求类型是否声明了 json 字段（即真正的请求体）
func hasBodyFields(route apispec.Route) bool {
	for _, param := range route.Params {
		if param.Location == apispec.LocationBody {
			return true
		}
	}
	return false
}

// hasPathFields 判断路由是否带 URL 路径参数
func hasPathFields(route apispec.Route) bool {
	for _, param := range route.Params {
		if param.Location == apispec.LocationPath {
			return true
		}
	}
	return false
}

// isFileUploadRoute 判断路由是否为文件上传（全部 form 字段且含 interface{} 类型）
func isFileUploadRoute(route apispec.Route) bool {
	if len(route.Params) == 0 {
		return false
	}
	hasFile := false
	for _, param := range route.Params {
		if param.Location != apispec.LocationForm {
			return false
		}
		if param.Type == "interface{}" {
			hasFile = true
		}
	}
	return hasFile
}

func extractBaseTypeName(typeName string) string {
	// 移除 []*  []  * 等前缀
	typeName = strings.TrimPrefix(typeName, "[]*")
	typeName = strings.TrimPrefix(typeName, "[]")
	typeName = strings.TrimPrefix(typeName, "*")
	return typeName
}

// generateIndexFile 生成 index.ts 文件，统一导出所有 API
func generateIndexFile(filePath string, apiExports []ApiExport) error {
	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	// 导出 types
	fmt.Fprintln(f, `export * from "./types";`)
	fmt.Fprintln(f)

	// 导出所有 API
	sort.Slice(apiExports, func(i, j int) bool {
		if apiExports[i].FileName != apiExports[j].FileName {
			return apiExports[i].FileName < apiExports[j].FileName
		}
		return apiExports[i].ApiName < apiExports[j].ApiName
	})
	for _, export := range apiExports {
		fmt.Fprintf(f, `export { %s } from "./%s";`+"\n", export.ApiName, export.FileName)
	}

	return nil
}
