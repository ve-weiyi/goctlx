package gin

import (
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/zeromicro/go-zero/tools/goctl/pkg/golang"
	"github.com/zeromicro/go-zero/tools/goctl/util/stringx"

	"github.com/zeromicro/go-zero/tools/goctl/util"

	"github.com/ve-weiyi/goctlx/x/gofile"

	"github.com/ve-weiyi/goctlx/parserx/apispec"
)

const (
	requestPkg  = "github.com/ve-weiyi/blog-gin/infra/request"
	responsePkg = "github.com/ve-weiyi/blog-gin/infra/response"
)

// generateAll orchestrates all code generation steps
func generateAll(service *apispec.ApiService, tplPath, outPath, contextPackage string) error {
	if err := generateTypes(service, tplPath, outPath); err != nil {
		return fmt.Errorf("generate types: %w", err)
	}
	if err := generateLogics(service, tplPath, outPath, contextPackage); err != nil {
		return fmt.Errorf("generate logics: %w", err)
	}
	if err := generateHandlers(service, tplPath, outPath, contextPackage); err != nil {
		return fmt.Errorf("generate handlers: %w", err)
	}
	if err := generateRouters(service, tplPath, outPath, contextPackage); err != nil {
		return fmt.Errorf("generate routers: %w", err)
	}
	if err := generateRoutes(service, tplPath, outPath, contextPackage); err != nil {
		return fmt.Errorf("generate routes: %w", err)
	}
	fmt.Println("✅ Gin code generated")
	return nil
}

func generateTypes(service *apispec.ApiService, tplPath, outPath string) error {
	tpl, err := os.ReadFile(path.Join(tplPath, "types.tpl"))
	if err != nil {
		return err
	}

	groupTypes := GroupTypes(service)

	t := util.With("types").Parse(string(tpl))
	for groupName, typeMap := range groupTypes {
		if len(typeMap) == 0 {
			continue
		}

		names := sortedTypeNames(typeMap)
		var types []string
		for _, name := range names {
			types = append(types, buildTypes(typeMap[name]))
		}

		// 展平目录：content/article → content_article.go
		flatName := strings.ReplaceAll(groupName, "/", "_")
		buf, err := t.Execute(map[string]any{
			"Package": "types",
			"Types":   types,
		})
		if err != nil {
			log.Println(err)
			continue
		}
		if err := gofile.Write(path.Join(outPath, "types", flatName+".go"), buf.Bytes()); err != nil {
			log.Println(err)
		}
	}
	return nil
}

func generateHandlers(service *apispec.ApiService, tplPath, outPath, contextPackage string) error {
	tpl, err := os.ReadFile(path.Join(tplPath, "handler.tpl"))
	if err != nil {
		return err
	}

	pkg, _, _ := golang.GetParentPackage(outPath)
	groups := ConvertRouteGroups(service)

	imports := []string{
		`"net/http"`,
		`"github.com/gin-gonic/gin"`,
		fmt.Sprintf(`"%s"`, contextPackage),
		fmt.Sprintf(`"%s/types"`, pkg),
		fmt.Sprintf(`"%s/logic"`, pkg),
		fmt.Sprintf(`"%s"`, requestPkg),
		fmt.Sprintf(`"%s"`, responsePkg),
	}

	t := util.With("handler").Parse(string(tpl))
	t.AddFunc("pkgTypes", pkgTypesFunc)
	t.AddFunc("commentTypes", commentTypesFunc)

	for k, v := range groups {
		buf, err := t.Execute(map[string]any{
			"Package":     "handler",
			"Imports":     imports,
			"Group":       k,
			"GroupRoutes": v,
		})
		if err != nil {
			log.Println(err)
			continue
		}
		if err := gofile.Write(path.Join(outPath, "handler", fmt.Sprintf("%v_handler.go", stringx.From(k).ToSnake())), buf.Bytes()); err != nil {
			log.Println(err)
		}
	}
	return nil
}

func generateLogics(service *apispec.ApiService, tplPath, outPath, contextPackage string) error {
	tpl, err := os.ReadFile(path.Join(tplPath, "logic.tpl"))
	if err != nil {
		return err
	}

	pkg, _, _ := golang.GetParentPackage(outPath)
	groups := ConvertRouteGroups(service)

	logicImports := []string{
		`"context"`,
		fmt.Sprintf(`"%s"`, contextPackage),
		fmt.Sprintf(`"%s/types"`, pkg),
		fmt.Sprintf(`"%s"`, requestPkg),
	}

	t := util.With("logic").Parse(string(tpl))
	t.AddFunc("pkgTypes", pkgTypesFunc)

	for k, v := range groups {
		buf, err := t.Execute(map[string]any{
			"Package":     "logic",
			"Imports":     logicImports,
			"Group":       k,
			"GroupRoutes": v,
		})
		if err != nil {
			log.Println(err)
			continue
		}
		if err := gofile.Write(path.Join(outPath, "logic", fmt.Sprintf("%v_logic.go", stringx.From(k).ToSnake())), buf.Bytes()); err != nil {
			log.Println(err)
		}
	}
	return nil
}

func generateRouters(service *apispec.ApiService, tplPath, outPath, contextPackage string) error {
	tpl, err := os.ReadFile(path.Join(tplPath, "router.tpl"))
	if err != nil {
		return err
	}

	pkg, _, _ := golang.GetParentPackage(outPath)
	groups := ConvertRouteGroups(service)

	t := util.With("router").Parse(string(tpl))

	for k, v := range groups {
		buf, err := t.Execute(map[string]any{
			"Package": "router",
			"Imports": []string{
				fmt.Sprintf(`"%s"`, contextPackage),
				fmt.Sprintf(`"%s/handler"`, pkg),
			},
			"Group":       k,
			"GroupRoutes": v,
		})
		if err != nil {
			log.Println(err)
			continue
		}
		if err := gofile.Write(path.Join(outPath, "router", fmt.Sprintf("%v_router.go", stringx.From(k).ToSnake())), buf.Bytes()); err != nil {
			log.Println(err)
		}
	}
	return nil
}

func generateRoutes(service *apispec.ApiService, tplPath, outPath, contextPackage string) error {
	tpl, err := os.ReadFile(path.Join(tplPath, "routes.tpl"))
	if err != nil {
		return err
	}

	pkg, _, _ := golang.GetParentPackage(outPath)
	groups := ConvertRouteGroups(service)

	gps := sortedGroups(groups)

	t := util.With("routes").Parse(string(tpl))
	t.AddFunc("Case2Camel", func(s string) string { return stringx.From(s).ToCamel() })
	t.AddFunc("Case2Snake", func(s string) string { return stringx.From(s).ToSnake() })
	t.AddFunc("ToUpper", strings.ToUpper)
	t.AddFunc("ToLower", strings.ToLower)

	buf, err := t.Execute(map[string]any{
		"Package": filepath.Base(outPath),
		"Imports": []string{
			fmt.Sprintf(`"%s"`, contextPackage),
			fmt.Sprintf(`"%s/router"`, pkg),
		},
		"Groups": gps,
	})
	if err != nil {
		return err
	}
	return gofile.Write(path.Join(outPath, "routes.go"), buf.Bytes())
}

// pkgTypesFunc 为类型名添加 types 包前缀，用于函数签名
func pkgTypesFunc(input string) string {
	if input == "" {
		return ""
	}
	re := regexp.MustCompile(`\w+`)
	result := re.ReplaceAllString(input, "types.$0")
	if strings.HasPrefix(result, "[]") {
		return result
	}
	return "*" + result
}

// commentTypesFunc 为类型名添加 types 包前缀（非指针），用于变量声明和 swagger 注释
func commentTypesFunc(input string) string {
	if input == "" {
		return ""
	}
	re := regexp.MustCompile(`\w+`)
	result := re.ReplaceAllString(input, "types.$0")
	return strings.ReplaceAll(result, "*", "")
}
