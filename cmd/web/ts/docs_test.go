package ts

import (
	"fmt"
	"testing"

	"github.com/go-openapi/loads"
	"github.com/ve-weiyi/pkg/utils/jsonconv"
	"github.com/zeromicro/go-zero/tools/goctl/api/parser"
)

const SWAGER_PATH = "/Users/weiyi/Github/veweiyi/goctlx/testdata/test.json"
const API_PATH = "/Users/weiyi/Github/sparkinai/sparkinai-cloud/service/api/app/proto/app.api"

func Test_Load(t *testing.T) {
	// Example with default loaders defined at the package level
	doc, err := loads.Spec(SWAGER_PATH)
	if err != nil {
		fmt.Println("Could not load this spec")
		return
	}

	sp := doc.Spec()

	t.Log(jsonconv.AnyToJsonIndent(sp))
}

func Test_Parser(t *testing.T) {
	parse, err := parser.Parse(API_PATH)
	if err != nil {
		return
	}

	t.Log(jsonconv.AnyToJsonIndent(parse))
}
