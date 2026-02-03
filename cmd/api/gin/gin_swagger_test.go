package gin

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ve-weiyi/goctlx/apispec"
)

func TestGenerateFromSwagger(t *testing.T) {
	// 测试文件路径
	swaggerFile := "../../../testdata/test.json"
	tplPath := "../../../template/api/gin"
	outPath := "../../../runtime/gin_swagger_test"
	contextPackage := "github.com/test/svctx"

	// 清理输出目录

	// 解析 swagger
	service, err := apispec.ParseSwaggerFromFile(swaggerFile)
	if err != nil {
		t.Fatalf("解析 swagger 失败: %v", err)
	}

	t.Logf("解析到 %d 个类型", len(service.Types))
	t.Logf("解析到 %d 个 API 分组", len(service.ApiGroups))

	// 测试生成 types
	if err := generateGinTypes(service, tplPath, outPath); err != nil {
		t.Errorf("生成 types 失败: %v", err)
	} else {
		checkFileExists(t, filepath.Join(outPath, "types"))
	}

	// 测试生成 logics
	if err := generateGinLogics(service, tplPath, outPath, contextPackage); err != nil {
		t.Errorf("生成 logics 失败: %v", err)
	} else {
		checkFileExists(t, filepath.Join(outPath, "logic"))
	}

	// 测试生成 handlers
	if err := generateGinHandlers(service, tplPath, outPath, contextPackage); err != nil {
		t.Errorf("生成 handlers 失败: %v", err)
	} else {
		checkFileExists(t, filepath.Join(outPath, "handler"))
	}

	// 测试生成 routers
	if err := generateGinRouters(service, tplPath, outPath, contextPackage); err != nil {
		t.Errorf("生成 routers 失败: %v", err)
	} else {
		checkFileExists(t, filepath.Join(outPath, "router"))
	}

	// 测试生成 routes
	if err := generateGinRoutes(service, tplPath, outPath, contextPackage); err != nil {
		t.Errorf("生成 routes 失败: %v", err)
	} else {
		checkFileExists(t, filepath.Join(outPath, "routes.go"))
	}
}

func checkFileExists(t *testing.T, path string) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("文件/目录不存在: %s", path)
	} else {
		t.Logf("✅ 生成成功: %s", path)
	}
}
