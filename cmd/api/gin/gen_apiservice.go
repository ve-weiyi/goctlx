package gin

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/zeromicro/go-zero/tools/goctl/pkg/golang"

	"github.com/ve-weiyi/goctlx/apispec"
	"github.com/ve-weiyi/pkg/kit/quickstart/gotplgen"
	"github.com/ve-weiyi/pkg/utils/jsonconv"
)

func generateTypesFromApiService(service *apispec.ApiService, tplPath, outPath string) error {
	tpl, err := os.ReadFile(path.Join(tplPath, "types.tpl"))
	if err != nil {
		return err
	}

	groupTypes := groupTypesByTag(service)
	var metas []gotplgen.TemplateMeta

	for groupName, types := range groupTypes {
		if len(types) == 0 {
			continue
		}

		var typeStrs []string
		for _, t := range types {
			typeStrs = append(typeStrs, buildTypeString(t))
		}

		meta := gotplgen.TemplateMeta{
			Mode:           gotplgen.ModeCreateOrReplace,
			CodeOutPath:    path.Join(outPath, "types", fmt.Sprintf("%v.go", groupName)),
			TemplateString: string(tpl),
			Data: map[string]any{
				"Package": "types",
				"Imports": []string{},
				"Name":    jsonconv.Case2Camel(groupName),
				"Types":   typeStrs,
			},
		}
		metas = append(metas, meta)
	}

	for _, m := range metas {
		if err := m.Execute(); err != nil {
			return err
		}
	}
	return nil
}

func generateLogicsFromApiService(service *apispec.ApiService, tplPath, outPath, contextPackage string) error {
	tpl, err := os.ReadFile(path.Join(tplPath, "logic.tpl"))
	if err != nil {
		return err
	}

	pkg, _, _ := golang.GetParentPackage(outPath)
	groups := convertApiServiceToGroups(service)
	var metas []gotplgen.TemplateMeta

	for groupName, groupRoutes := range groups {
		meta := gotplgen.TemplateMeta{
			Mode:           gotplgen.ModeCreateOrReplace,
			CodeOutPath:    path.Join(outPath, "logic", fmt.Sprintf("%v_logic.go", groupName)),
			TemplateString: string(tpl),
			Data: map[string]any{
				"Package": "logic",
				"Imports": []string{
					fmt.Sprintf(`"%s"`, contextPackage),
					fmt.Sprintf(`"%s/types"`, pkg),
				},
				"Group":       jsonconv.Case2Camel(groupName),
				"GroupRoutes": groupRoutes,
			},
			FunMap: map[string]any{
				"pkgTypes": func(input string) string {
					if input == "" {
						return ""
					}
					return "*types." + input
				},
			},
		}
		metas = append(metas, meta)
	}

	for _, m := range metas {
		if err := m.Execute(); err != nil {
			return err
		}
	}
	return nil
}

func generateHandlersFromApiService(service *apispec.ApiService, tplPath, outPath, contextPackage string) error {
	tpl, err := os.ReadFile(path.Join(tplPath, "handler.tpl"))
	if err != nil {
		return err
	}

	pkg, _, _ := golang.GetParentPackage(outPath)
	groups := convertApiServiceToGroups(service)
	var metas []gotplgen.TemplateMeta

	for groupName, groupRoutes := range groups {
		meta := gotplgen.TemplateMeta{
			Mode:           gotplgen.ModeCreateOrReplace,
			CodeOutPath:    path.Join(outPath, "handler", fmt.Sprintf("%v_handler.go", groupName)),
			TemplateString: string(tpl),
			Data: map[string]any{
				"Package": "handler",
				"Imports": []string{
					fmt.Sprintf(`"%s"`, contextPackage),
					fmt.Sprintf(`"%s/types"`, pkg),
					fmt.Sprintf(`"%s/logic"`, pkg),
				},
				"Group":       jsonconv.Case2Camel(groupName),
				"GroupRoutes": groupRoutes,
			},
			FunMap: map[string]any{
				"pkgTypes": func(input string) string {
					if input == "" {
						return ""
					}
					return "*types." + input
				},
				"commentTypes": func(input string) string {
					if input == "" {
						return ""
					}
					return "types." + input
				},
			},
		}
		metas = append(metas, meta)
	}

	for _, m := range metas {
		if err := m.Execute(); err != nil {
			return err
		}
	}
	return nil
}

func generateRoutersFromApiService(service *apispec.ApiService, tplPath, outPath, contextPackage string) error {
	tpl, err := os.ReadFile(path.Join(tplPath, "router.tpl"))
	if err != nil {
		return err
	}

	pkg, _, _ := golang.GetParentPackage(outPath)
	groups := convertApiServiceToGroups(service)
	var metas []gotplgen.TemplateMeta

	for groupName, groupRoutes := range groups {
		meta := gotplgen.TemplateMeta{
			Mode:           gotplgen.ModeCreateOrReplace,
			CodeOutPath:    path.Join(outPath, "router", fmt.Sprintf("%v_router.go", groupName)),
			TemplateString: string(tpl),
			Data: map[string]any{
				"Package": "router",
				"Imports": []string{
					fmt.Sprintf(`"%s"`, contextPackage),
					fmt.Sprintf(`"%s/types"`, pkg),
					fmt.Sprintf(`"%s/handler"`, pkg),
				},
				"Group":       jsonconv.Case2Camel(groupName),
				"GroupRoutes": groupRoutes,
			},
		}
		metas = append(metas, meta)
	}

	for _, m := range metas {
		if err := m.Execute(); err != nil {
			return err
		}
	}
	return nil
}

func generateRoutesFromApiService(service *apispec.ApiService, tplPath, outPath, contextPackage string) error {
	tpl, err := os.ReadFile(path.Join(tplPath, "routes.tpl"))
	if err != nil {
		return err
	}

	pkg, _, _ := golang.GetParentPackage(outPath)
	groups := convertApiServiceToGroups(service)

	var groupNames []string
	for groupName := range groups {
		groupNames = append(groupNames, jsonconv.Case2Camel(groupName))
	}
	sort.Strings(groupNames)

	meta := gotplgen.TemplateMeta{
		Mode:           gotplgen.ModeCreateOrReplace,
		CodeOutPath:    path.Join(outPath, "routes.go"),
		TemplateString: string(tpl),
		Data: map[string]any{
			"Package": filepath.Base(outPath),
			"Imports": []string{
				fmt.Sprintf(`"%s"`, contextPackage),
				fmt.Sprintf(`"%s/router"`, pkg),
			},
			"Groups": groupNames,
		},
		FunMap: gotplgen.StdMapUtils,
	}

	return meta.Execute()
}

func convertApiServiceToGroups(service *apispec.ApiService) map[string][]GroupRoute {
	groups := make(map[string][]GroupRoute)

	for _, apiGroup := range service.ApiGroups {
		groupName := safeGroupName(apiGroup.Prefix)

		var routes []Route
		for _, r := range apiGroup.Routes {
			routes = append(routes, Route{
				Doc:      r.Summary,
				Handler:  jsonconv.Case2Camel(safeIdentifier(r.Handler)),
				Path:     r.Path,
				Method:   r.Method,
				Request:  r.Request,
				Response: r.Response,
			})
		}

		if len(routes) == 0 {
			continue
		}

		gr := GroupRoute{
			Name:       groupName,
			Prefix:     apiGroup.Prefix,
			Middleware: apiGroup.Middleware,
			Routes:     routes,
		}

		groups[groupName] = append(groups[groupName], gr)
	}

	return groups
}

func safeGroupName(name string) string {
	if name == "" {
		return "default"
	}
	name = strings.ToLower(name)
	return safeIdentifier(name)
}

func safeIdentifier(s string) string {
	if s == "" {
		return "default"
	}
	var result strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' {
			result.WriteRune(r)
		}
	}
	if result.Len() == 0 {
		return "default"
	}
	return result.String()
}

func groupTypesByTag(service *apispec.ApiService) map[string][]apispec.Type {
	typeUsage := make(map[string]map[string]bool)

	for _, group := range service.ApiGroups {
		tag := safeGroupName(group.Prefix)

		for _, route := range group.Routes {
			if route.Request != "" {
				if typeUsage[route.Request] == nil {
					typeUsage[route.Request] = make(map[string]bool)
				}
				typeUsage[route.Request][tag] = true
			}
			if route.Response != "" {
				if typeUsage[route.Response] == nil {
					typeUsage[route.Response] = make(map[string]bool)
				}
				typeUsage[route.Response][tag] = true
			}
		}
	}

	result := make(map[string][]apispec.Type)
	for _, t := range service.Types {
		tags := typeUsage[t.Name]
		var targetTag string

		if len(tags) == 1 {
			for tag := range tags {
				targetTag = tag
			}
		} else {
			targetTag = "types"
		}

		result[targetTag] = append(result[targetTag], t)
	}

	for tag := range result {
		sort.Slice(result[tag], func(i, j int) bool {
			return result[tag][i].Name < result[tag][j].Name
		})
	}

	return result
}

func buildTypeString(t apispec.Type) string {
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("type %s struct {\n", jsonconv.Case2Camel(t.Name)))

	for _, field := range t.Fields {
		builder.WriteString(fmt.Sprintf("\t%s %s %s",
			jsonconv.Case2Camel(field.Name),
			field.Type,
			field.Tag))

		if field.Comment != "" {
			builder.WriteString(fmt.Sprintf(" // %s", field.Comment))
		}
		builder.WriteString("\n")
	}

	builder.WriteString("}")
	return builder.String()
}
