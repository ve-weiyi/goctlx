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

	return generateFromApiService(typescriptApiFlags.OutPath, convertApiSpecToService(apiSpec))
}

func generateFromApiService(outPath string, apiData *apispec.ApiService) error {
	if err := cleanOutputDir(outPath); err != nil {
		return fmt.Errorf("failed to clean output directory: %w", err)
	}

	if err := os.MkdirAll(outPath, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	typesFile := filepath.Join(outPath, "types.ts")
	if err := generateTypesFile(typesFile, apiData); err != nil {
		return fmt.Errorf("failed to generate types file: %w", err)
	}
	fmt.Printf("✅ Generated: %s\n", typesFile)

	var apiExports []ApiExport
	for _, group := range apiData.Groups {
		safeFileName := ConvertPathToKebabCase(group.Name)
		if safeFileName == "" {
			safeFileName = strings.ToLower(group.Prefix)
		}
		apiFile := filepath.Join(outPath, safeFileName+".ts")
		if dir := filepath.Dir(apiFile); dir != outPath {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return fmt.Errorf("failed to create directory: %w", err)
			}
		}
		if err := generateApiFile(apiFile, group); err != nil {
			return fmt.Errorf("failed to generate api file: %w", err)
		}
		fmt.Printf("✅ Generated: %s\n", apiFile)
		apiExports = append(apiExports, ApiExport{
			FileName: safeFileName,
			ApiName:  LastSegmentPascalCase(group.Name) + "API",
		})
	}

	indexFile := filepath.Join(outPath, "index.ts")
	if err := generateIndexFile(indexFile, apiExports); err != nil {
		return fmt.Errorf("failed to generate index file: %w", err)
	}
	fmt.Printf("✅ Generated: %s\n", indexFile)

	fmt.Println("TypeScript code generated successfully")
	return nil
}

// cleanOutputDir 生成前清空输出目录。
// 契约换域（如 permission/ → access/）后旧文件不会被覆盖，留在原处会让前端
// 同时存在新旧两套 URL；goctl 同样不清理残留文件，这里由本工具负责。
// 目录里已有本工具生成的 index.ts 才允许清理，避免 -o 指错目录时误删。
func cleanOutputDir(outPath string) error {
	entries, err := os.ReadDir(outPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if len(entries) == 0 {
		return nil
	}
	if _, err := os.Stat(filepath.Join(outPath, "index.ts")); err != nil {
		return fmt.Errorf("输出目录 %s 中未见本工具生成的 index.ts，已中止以免误删", outPath)
	}
	for _, entry := range entries {
		if err := os.RemoveAll(filepath.Join(outPath, entry.Name())); err != nil {
			return err
		}
	}
	return nil
}

func convertApiSpecToService(apiSpec *spec.ApiSpec) *apispec.ApiService {
	service := &apispec.ApiService{
		Name:   apiSpec.Info.Properties["title"],
		Types:  []apispec.Type{},
		Groups: []apispec.ApiGroup{},
	}

	// 建立类型名 → 字段列表的索引
	typeFields := make(map[string][]apispec.Field)

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
					// 将内联类型的字段也加入
					if inherited, ok := typeFields[member.Type.Name()]; ok {
						t.Fields = append(t.Fields, inherited...)
					}
					continue
				}
				field := parseField(member)
				t.Fields = append(t.Fields, field)
			}
			t.Extends = inlineTypes
			typeFields[t.Name] = t.Fields
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
				Label:      group.GetAnnotation("tags"),
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
				Handler: handler,
				Summary: route.AtDoc.Text,
				Path:    route.Path,
				Method:  route.Method,
			}

			if route.RequestType != nil {
				reqName := route.RequestType.Name()
				r.Request = reqName
				r.Params = typeFields[reqName]
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
		service.Groups = append(service.Groups, *group)
	}

	return service
}

func parseField(member spec.Member) apispec.Field {
	field := apispec.Field{
		Name:     member.Name,
		Type:     member.Type.Name(),
		Location: apispec.LocationBody,
		Comment:  member.Comment,
	}

	if member.Tag == "" {
		return field
	}

	tagTypes := []struct {
		key string
		loc apispec.ParamLocation
	}{
		{"json", apispec.LocationBody},
		{"form", apispec.LocationForm},
		{"path", apispec.LocationPath},
		{"header", apispec.LocationHeader},
	}

	for _, tt := range tagTypes {
		name, opts := parseTagValue(member.Tag, tt.key)
		if name != "" {
			field.Name = name
			field.Location = tt.loc
			field.Optional = containsWord(opts, "optional")
			return field
		}
	}
	return field
}

// parseTagValue 从 tag 字符串中提取指定 key 的值和选项
// 例如: parseTagValue("`json:\"name,optional\"`", "json") → ("name", ["optional"])
func parseTagValue(tag, key string) (string, []string) {
	prefix := key + ":\""
	start := strings.Index(tag, prefix)
	if start == -1 {
		return "", nil
	}
	start += len(prefix)
	end := strings.Index(tag[start:], "\"")
	if end == -1 {
		return "", nil
	}
	value := tag[start : start+end]
	parts := strings.Split(value, ",")
	if len(parts) == 0 {
		return "", nil
	}
	return parts[0], parts[1:]
}

func containsWord(opts []string, word string) bool {
	for _, opt := range opts {
		if strings.TrimSpace(opt) == word {
			return true
		}
	}
	return false
}
