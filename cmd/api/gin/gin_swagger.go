package gin

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ve-weiyi/goctlx/parserx/apispec"
)

var ginSwaggerFlags = struct {
	SwaggerFile    string
	TplPath        string
	OutPath        string
	ContextPackage string
}{
	SwaggerFile:    "swagger.json",
	TplPath:        "./template/api/gin",
	OutPath:        "./output",
	ContextPackage: "github.com/example/svctx",
}

func NewGinSwaggerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "swagger",
		Short: "从 swagger.json 文件生成 Gin 框架代码",
		RunE:  runGinSwagger,
	}

	cmd.Flags().StringVarP(&ginSwaggerFlags.SwaggerFile, "api-file", "f", ginSwaggerFlags.SwaggerFile, "swagger.json 文件路径")
	cmd.Flags().StringVarP(&ginSwaggerFlags.TplPath, "tpl-path", "t", ginSwaggerFlags.TplPath, "模板目录路径")
	cmd.Flags().StringVarP(&ginSwaggerFlags.OutPath, "out-path", "o", ginSwaggerFlags.OutPath, "输出目录路径")
	cmd.Flags().StringVarP(&ginSwaggerFlags.ContextPackage, "svctx-package", "c", ginSwaggerFlags.ContextPackage, "svctx 包导入路径")

	return cmd
}

func runGinSwagger(cmd *cobra.Command, args []string) error {
	fmt.Println("===== 命令参数 =====")
	fmt.Printf("swagger-file: %s\n", ginSwaggerFlags.SwaggerFile)
	fmt.Printf("tpl-path: %s\n", ginSwaggerFlags.TplPath)
	fmt.Printf("out-path: %s\n", ginSwaggerFlags.OutPath)
	fmt.Printf("svctx-package: %s\n", ginSwaggerFlags.ContextPackage)
	fmt.Println("====================")

	service, err := apispec.ParseSwaggerFromFile(ginSwaggerFlags.SwaggerFile)
	if err != nil {
		return fmt.Errorf("failed to parse swagger file: %w", err)
	}

	return generateAll(service, ginSwaggerFlags.TplPath, ginSwaggerFlags.OutPath, ginSwaggerFlags.ContextPackage)
}
