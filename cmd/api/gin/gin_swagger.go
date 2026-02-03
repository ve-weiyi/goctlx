package gin

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ve-weiyi/goctlx/apispec"
)

var swagFlags = struct {
	SwaggerFile    string
	TplPath        string
	OutPath        string
	ContextPackage string
}{
	SwaggerFile:    "test.json",
	TplPath:        "./template/api/gin",
	OutPath:        "./",
	ContextPackage: "context",
}

func NewGinSwaggerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "swagger",
		Short: "从 swagger.json 生成 Gin 框架代码",
		RunE:  runSwagger,
	}

	cmd.Flags().StringVarP(&swagFlags.SwaggerFile, "api-file", "f", swagFlags.SwaggerFile, "Swagger文件路径")
	cmd.Flags().StringVarP(&swagFlags.TplPath, "tpl-path", "t", swagFlags.TplPath, "模板路径")
	cmd.Flags().StringVarP(&swagFlags.OutPath, "out-path", "o", swagFlags.OutPath, "输出路径")
	cmd.Flags().StringVarP(&swagFlags.ContextPackage, "svctx-package", "c", swagFlags.ContextPackage, "上下文包名")

	return cmd
}

func runSwagger(cmd *cobra.Command, args []string) error {
	fmt.Println("===== 命令参数 =====")
	fmt.Printf("swagger-file: %s\n", swagFlags.SwaggerFile)
	fmt.Printf("tpl-path: %s\n", swagFlags.TplPath)
	fmt.Printf("out-path: %s\n", swagFlags.OutPath)
	fmt.Printf("context-package: %s\n", swagFlags.ContextPackage)
	fmt.Println("====================")

	// 解析 swagger.json
	apiService, err := apispec.ParseSwaggerFromFile(swagFlags.SwaggerFile)
	if err != nil {
		return err
	}

	// 生成代码
	if err := generateGinTypes(apiService, swagFlags.TplPath, swagFlags.OutPath); err != nil {
		return fmt.Errorf("failed to generate types: %w", err)
	}
	fmt.Println("✅ Generated types")

	if err := generateGinLogics(apiService, swagFlags.TplPath, swagFlags.OutPath, swagFlags.ContextPackage); err != nil {
		return fmt.Errorf("failed to generate logics: %w", err)
	}
	fmt.Println("✅ Generated logics")

	if err := generateGinHandlers(apiService, swagFlags.TplPath, swagFlags.OutPath, swagFlags.ContextPackage); err != nil {
		return fmt.Errorf("failed to generate handlers: %w", err)
	}
	fmt.Println("✅ Generated handlers")

	if err := generateGinRouters(apiService, swagFlags.TplPath, swagFlags.OutPath, swagFlags.ContextPackage); err != nil {
		return fmt.Errorf("failed to generate routers: %w", err)
	}
	fmt.Println("✅ Generated routers")

	if err := generateGinRoutes(apiService, swagFlags.TplPath, swagFlags.OutPath, swagFlags.ContextPackage); err != nil {
		return fmt.Errorf("failed to generate routes: %w", err)
	}
	fmt.Println("✅ Generated routes")

	return nil
}

// ============ 代码生成实现函数 ============

func generateGinTypes(service *apispec.ApiService, tplPath, outPath string) error {
	return generateTypesFromApiService(service, tplPath, outPath)
}

func generateGinLogics(service *apispec.ApiService, tplPath, outPath, contextPackage string) error {
	return generateLogicsFromApiService(service, tplPath, outPath, contextPackage)
}

func generateGinHandlers(service *apispec.ApiService, tplPath, outPath, contextPackage string) error {
	return generateHandlersFromApiService(service, tplPath, outPath, contextPackage)
}

func generateGinRouters(service *apispec.ApiService, tplPath, outPath, contextPackage string) error {
	return generateRoutersFromApiService(service, tplPath, outPath, contextPackage)
}

func generateGinRoutes(service *apispec.ApiService, tplPath, outPath, contextPackage string) error {
	return generateRoutesFromApiService(service, tplPath, outPath, contextPackage)
}
