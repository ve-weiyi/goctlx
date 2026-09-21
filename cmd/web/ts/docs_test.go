package ts

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/go-openapi/loads"
	"github.com/zeromicro/go-zero/tools/goctl/api/parser"
)

// 以下路径来自开发机，仓库内没有对应文件；不存在时测试跳过
const SWAGER_PATH = "/Users/weiyi/Github/veweiyi/goctlx/testdata/test.json"
const API_PATH = "/Users/weiyi/Github/blog/service/app/api/proto/app.api"

func skipIfAbsent(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err != nil {
		t.Skipf("跳过：%s 不存在", path)
	}
}

func Test_Load(t *testing.T) {
	skipIfAbsent(t, SWAGER_PATH)

	// Example with default loaders defined at the package level
	doc, err := loads.Spec(SWAGER_PATH)
	if err != nil {
		fmt.Println("Could not load this spec")
		return
	}

	sp := doc.Spec()

	t.Log(dumpJSON(sp))
}

func Test_Parser(t *testing.T) {
	// goctl 的 parser 对不存在的文件不返回 error，而是内部 log.Fatal 杀掉测试进程，
	// 因此必须先自行判断文件是否存在
	skipIfAbsent(t, API_PATH)

	parse, err := parser.Parse(API_PATH)
	if err != nil {
		return
	}

	t.Log(dumpJSON(parse))
}

// dumpJSON 把值序列化为缩进 JSON，便于测试输出查看
func dumpJSON(v any) string {
	data, err := json.MarshalIndent(v, "", " ")
	if err != nil || string(data) == "{}" {
		return fmt.Sprintf("%+v", v)
	}
	return string(data)
}
