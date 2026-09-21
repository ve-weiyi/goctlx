package gin

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/zeromicro/go-zero/tools/goctl/api/parser"
	"github.com/zeromicro/go-zero/tools/goctl/api/spec"

	"github.com/ve-weiyi/goctlx/parserx/apispec"
)

var ginApiFlags = struct {
	ApiFile        string
	TplPath        string
	OutPath        string
	ContextPackage string
}{
	ApiFile:        "app.api",
	TplPath:        "./template/api/gin",
	OutPath:        "./output",
	ContextPackage: "github.com/example/svctx",
}

func NewGinApiCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "api",
		Short: "从 .api 文件生成 Gin 框架代码",
		RunE:  runGinApi,
	}

	cmd.Flags().StringVarP(&ginApiFlags.ApiFile, "api-file", "f", ginApiFlags.ApiFile, ".api 文件路径")
	cmd.Flags().StringVarP(&ginApiFlags.TplPath, "tpl-path", "t", ginApiFlags.TplPath, "模板目录路径")
	cmd.Flags().StringVarP(&ginApiFlags.OutPath, "out-path", "o", ginApiFlags.OutPath, "输出目录路径")
	cmd.Flags().StringVarP(&ginApiFlags.ContextPackage, "svctx-package", "c", ginApiFlags.ContextPackage, "svctx 包导入路径")

	return cmd
}

func runGinApi(cmd *cobra.Command, args []string) error {
	fmt.Println("===== 命令参数 =====")
	fmt.Printf("api-file: %s\n", ginApiFlags.ApiFile)
	fmt.Printf("tpl-path: %s\n", ginApiFlags.TplPath)
	fmt.Printf("out-path: %s\n", ginApiFlags.OutPath)
	fmt.Printf("svctx-package: %s\n", ginApiFlags.ContextPackage)
	fmt.Println("====================")

	apiSpec, err := parser.Parse(ginApiFlags.ApiFile)
	if err != nil {
		return fmt.Errorf("failed to parse api file: %w", err)
	}

	service := convertGoZeroSpec(apiSpec)
	return generateAll(service, ginApiFlags.TplPath, ginApiFlags.OutPath, ginApiFlags.ContextPackage)
}

// convertGoZeroSpec 将 go-zero ApiSpec 转换为 goctlx ApiService
func convertGoZeroSpec(apiSpec *spec.ApiSpec) *apispec.ApiService {
	service := &apispec.ApiService{
		Name:   "api",
		Types:  []apispec.Type{},
		Groups: []apispec.ApiGroup{},
	}

	typeFields := make(map[string][]apispec.Field)

	for _, typ := range apiSpec.Types {
		defStruct, ok := typ.(spec.DefineStruct)
		if !ok {
			continue
		}
		t := apispec.Type{
			Name:    defStruct.Name(),
			Comment: "",
			Fields:  []apispec.Field{},
		}
		if len(defStruct.Docs) > 0 {
			t.Comment = defStruct.Docs[0]
		}

		var inlineTypes []string
		for _, member := range defStruct.Members {
			if member.IsInline {
				inlineTypes = append(inlineTypes, member.Type.Name())
				if inherited, ok := typeFields[member.Type.Name()]; ok {
					t.Fields = append(t.Fields, inherited...)
				}
				continue
			}
			field := parseMemberField(member)
			t.Fields = append(t.Fields, field)
		}
		t.Extends = inlineTypes
		typeFields[t.Name] = t.Fields
		service.Types = append(service.Types, t)
	}

	for _, group := range apiSpec.Service.Groups {
		groupName := group.GetAnnotation("group")
		if groupName == "" {
			groupName = "default"
		}

		var middleware []string
		mw := group.GetAnnotation("middleware")
		if mw != "" {
			middleware = strings.Split(mw, ",")
		}

		apiGroup := apispec.ApiGroup{
			Name:       groupName,
			Prefix:     group.GetAnnotation("prefix"),
			Label:      group.GetAnnotation("tags"),
			Middleware: middleware,
			Routes:     []apispec.Route{},
		}

		for _, route := range group.Routes {
			handler := route.Handler
			if len(handler) > 0 {
				handler = strings.ToLower(handler[:1]) + handler[1:]
			}

			r := apispec.Route{
				Handler: handler,
				Summary: route.AtDoc.Text,
				Path:    route.Path,
				Method:  route.Method,
			}

			if route.RequestType != nil {
				r.Request = route.RequestType.Name()
				r.Params = typeFields[route.RequestType.Name()]
			}
			if route.ResponseType != nil {
				r.Response = route.ResponseType.Name()
			}

			apiGroup.Routes = append(apiGroup.Routes, r)
		}

		service.Groups = append(service.Groups, apiGroup)
	}

	return service
}

// parseMemberField 解析 go-zero spec.Member 为 apispec.Field
func parseMemberField(member spec.Member) apispec.Field {
	field := apispec.Field{
		Name:     member.Name,
		Type:     member.Type.Name(),
		Location: apispec.LocationBody,
		Comment:  member.Comment,
		Optional: false,
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
	if len(parts) > 1 {
		return parts[0], parts[1:]
	}
	return parts[0], nil
}

func containsWord(opts []string, word string) bool {
	for _, opt := range opts {
		if strings.TrimSpace(opt) == word {
			return true
		}
	}
	return false
}
