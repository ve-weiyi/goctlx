package ts

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/zeromicro/go-zero/tools/goctl/api/parser"
	"github.com/zeromicro/go-zero/tools/goctl/api/spec"

	"github.com/ve-weiyi/goctlx/parserx/apispec"
)

var typescriptApiFlags = struct {
	ApiFile string
	OutPath string
}{
	ApiFile: "test.api",
	OutPath: "./",
}

func NewTypescriptApiCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "api",
		Short: "根据 .api 文件生成 TypeScript 代码",
		RunE:  runTypescriptApi,
	}

	cmd.Flags().StringVarP(&typescriptApiFlags.ApiFile, "api-file", "f", typescriptApiFlags.ApiFile, "API文件路径")
	cmd.Flags().StringVarP(&typescriptApiFlags.OutPath, "out-path", "o", typescriptApiFlags.OutPath, "输出目录路径")

	return cmd
}

func runTypescriptApi(cmd *cobra.Command, args []string) error {
	fmt.Println("===== 命令参数 =====")
	fmt.Printf("api-file: %s\n", typescriptApiFlags.ApiFile)
	fmt.Printf("out-path: %s\n", typescriptApiFlags.OutPath)
	fmt.Println("====================")

	apiSpec, err := parser.Parse(typescriptApiFlags.ApiFile)
	if err != nil {
		return fmt.Errorf("failed to parse api file: %w", err)
	}

	apiService := convertApiSpecToService(apiSpec)

	if err = os.MkdirAll(typescriptApiFlags.OutPath, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	if err = generateTypesFile(typescriptApiFlags.OutPath+"/types.ts", apiService); err != nil {
		return fmt.Errorf("failed to generate types file: %w", err)
	}
	fmt.Printf("✅ Generated: %s/types.ts\n", typescriptApiFlags.OutPath)

	// 收集所有 API 导出信息
	var apiExports []ApiExport

	for _, group := range apiService.ApiGroups {
		// 使用辅助函数将路径转换为 snake_case 文件名
		safeFileName := ConvertPathToSnakeCase(group.Name)
		fileName := fmt.Sprintf("%s/%s.ts", typescriptApiFlags.OutPath, safeFileName)
		fileDir := filepath.Dir(fileName)
		if err = os.MkdirAll(fileDir, 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
		if err = generateApiFile(fileName, group); err != nil {
			return fmt.Errorf("failed to generate api file: %w", err)
		}
		fmt.Printf("✅ Generated: %s\n", fileName)

		// 收集导出信息
		apiName := ConvertPathToPascalCase(group.Name) + "API"
		apiExports = append(apiExports, ApiExport{
			FileName: safeFileName,
			ApiName:  apiName,
		})
	}

	// 生成 index.ts 文件
	indexFile := filepath.Join(typescriptApiFlags.OutPath, "index.ts")
	if err = generateIndexFile(indexFile, apiExports); err != nil {
		return fmt.Errorf("failed to generate index file: %w", err)
	}
	fmt.Printf("✅ Generated: %s\n", indexFile)

	fmt.Println("TypeScript code generated successfully")
	return nil
}

func convertApiSpecToService(apiSpec *spec.ApiSpec) *apispec.ApiService {
	service := &apispec.ApiService{
		Name:      apiSpec.Info.Properties["title"],
		Types:     []apispec.Type{},
		ApiGroups: []apispec.ApiGroup{},
	}

	for _, typ := range apiSpec.Types {
		if defStruct, ok := typ.(spec.DefineStruct); ok {
			t := apispec.Type{
				Name:   defStruct.Name(),
				Fields: []apispec.Field{},
			}
			if len(defStruct.Docs) > 0 {
				t.Comment = defStruct.Docs[0]
			}

			var inlineTypes []string
			for _, member := range defStruct.Members {
				if member.IsInline {
					inlineTypes = append(inlineTypes, member.Type.Name())
					continue
				}
				fieldName := member.Name
				if member.Tag != "" {
					if jsonName := extractJsonName(member.Tag); jsonName != "" {
						fieldName = jsonName
					}
				}
				// 检查是否为 optional
				isOptional := strings.Contains(member.Tag, "optional")
				field := apispec.Field{
					Name:     fieldName,
					Type:     member.Type.Name(),
					Tag:      member.Tag,
					Comment:  member.Comment,
					Nullable: isOptional,
				}
				t.Fields = append(t.Fields, field)
			}
			t.Extends = inlineTypes
			service.Types = append(service.Types, t)
		}
	}

	groupMap := make(map[string]*apispec.ApiGroup)
	for _, group := range apiSpec.Service.Groups {
		groupName := group.GetAnnotation("group")
		if groupName == "" {
			groupName = "default"
		}

		if groupMap[groupName] == nil {
			groupMap[groupName] = &apispec.ApiGroup{
				Name:       groupName,
				Prefix:     group.GetAnnotation("prefix"),
				Tag:        group.GetAnnotation("tags"),
				Middleware: []string{},
				Routes:     []apispec.Route{},
			}
		}

		for _, route := range group.Routes {
			handler := route.Handler
			if handler != "" {
				handler = strings.ToLower(handler[:1]) + handler[1:]
			}
			r := apispec.Route{
				Handler:     handler,
				Summary:     route.AtDoc.Text,
				Path:        route.Path,
				Method:      route.Method,
				QueryParams: []apispec.QueryParam{},
			}

			if route.RequestType != nil {
				r.Request = route.RequestType.Name()
			}
			if route.ResponseType != nil {
				r.Response = route.ResponseType.Name()
			}

			groupMap[groupName].Routes = append(groupMap[groupName].Routes, r)
		}
	}

	for groupName, group := range groupMap {
		if group.Prefix == "" {
			group.Prefix = groupName
		}
		service.ApiGroups = append(service.ApiGroups, *group)
	}

	return service
}

func extractJsonName(tag string) string {
	if tag == "" {
		return ""
	}

	// 优先使用 json tag
	if start := strings.Index(tag, "json:\""); start != -1 {
		start += 6
		if end := strings.Index(tag[start:], "\""); end != -1 {
			jsonTag := tag[start : start+end]
			if idx := strings.Index(jsonTag, ","); idx != -1 {
				jsonTag = jsonTag[:idx]
			}
			return jsonTag
		}
	}

	// 其次使用 form tag
	if start := strings.Index(tag, "form:\""); start != -1 {
		start += 6
		if end := strings.Index(tag[start:], "\""); end != -1 {
			formTag := tag[start : start+end]
			if idx := strings.Index(formTag, ","); idx != -1 {
				formTag = formTag[:idx]
			}
			return formTag
		}
	}

	// 最后使用 path tag
	if start := strings.Index(tag, "path:\""); start != -1 {
		start += 6
		if end := strings.Index(tag[start:], "\""); end != -1 {
			pathTag := tag[start : start+end]
			if idx := strings.Index(pathTag, ","); idx != -1 {
				pathTag = pathTag[:idx]
			}
			return pathTag
		}
	}

	return ""
}
