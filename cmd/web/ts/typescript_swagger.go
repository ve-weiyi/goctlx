package ts

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ve-weiyi/goctlx/apispec"
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
	apiData, err := apispec.ParseSwaggerFromFile(typescriptSwaggerFlags.ApiFile)
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

	// 为每个分组生成独立的 API 文件
	for _, group := range apiData.ApiGroups {
		fileName := fmt.Sprintf("%s.ts", strings.ToLower(group.Prefix))
		outputFile := filepath.Join(typescriptSwaggerFlags.OutPath, fileName)
		if err := generateApiFile(outputFile, group); err != nil {
			return fmt.Errorf("failed to generate api file: %w", err)
		}
		fmt.Printf("✅ Generated: %s\n", outputFile)
	}

	return nil
}

// ============ 代码生成 ============

func generateTypesFile(filePath string, data *apispec.ApiService) error {
	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	for _, t := range data.Types {
		if t.Comment != "" {
			fmt.Fprintf(f, "// %s\n", t.Comment)
		}
		fmt.Fprintf(f, "export interface %s {\n", t.Name)
		for _, field := range t.Fields {
			nullable := ""
			if field.Nullable {
				nullable = "?"
			}
			comment := ""
			if field.Comment != "" {
				comment = " // " + field.Comment
			}
			tsType := ConvertGoTypeToTsType(field.Type)
			fmt.Fprintf(f, "  %s%s: %s;%s\n", field.Name, nullable, tsType, comment)
		}
		fmt.Fprintln(f, "}")
		fmt.Fprintln(f)
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
			typeSet[route.Request] = true
		}
		if route.Response != "" {
			typeSet[route.Response] = true
		}
	}

	// 生成类型导入
	if len(typeSet) > 0 {
		fmt.Fprint(f, `import type { `)
		i := 0
		for typeName := range typeSet {
			if i > 0 {
				fmt.Fprint(f, ", ")
			}
			fmt.Fprint(f, typeName)
			i++
		}
		fmt.Fprintln(f, ` } from "./types";`)
	}
	fmt.Fprintln(f)

	// API 对象
	groupName := strings.Title(group.Prefix)
	if groupName == "" {
		groupName = "Default"
	}

	if group.Tag != "" {
		fmt.Fprintf(f, "/** %s */\n", group.Tag)
	}
	fmt.Fprintf(f, "export const %sAPI = {\n", groupName)

	for _, route := range group.Routes {
		generateApiMethod(f, route)
	}

	fmt.Fprintln(f, "};")
	return nil
}

func generateApiMethod(f *os.File, route apispec.Route) {
	if route.Summary != "" {
		fmt.Fprintf(f, "  /** %s */\n", route.Summary)
	}

	// 函数参数
	var params []string
	if route.Request != "" {
		params = append(params, fmt.Sprintf("data?: %s", route.Request))
	}
	if len(route.QueryParams) > 0 {
		params = append(params, "params?: Record<string, any>")
	}
	reqParam := strings.Join(params, ", ")

	response := "any"
	if route.Response != "" {
		response = route.Response
	}

	fmt.Fprintf(f, "  %s(%s): Promise<IApiResponse<%s>> {\n", route.Handler, reqParam, response)

	// Request 调用
	fmt.Fprintln(f, "    return request({")
	fmt.Fprintf(f, "      url: \"%s\",\n", route.Path)
	fmt.Fprintf(f, "      method: \"%s\",\n", route.Method)

	if route.Request != "" {
		fmt.Fprintln(f, "      data: data,")
	}
	if len(route.QueryParams) > 0 {
		fmt.Fprintln(f, "      params: params,")
	}

	fmt.Fprintln(f, "    });")
	fmt.Fprintln(f, "  },")
	fmt.Fprintln(f)
}
