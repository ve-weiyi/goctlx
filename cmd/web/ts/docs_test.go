package ts

import (
	"fmt"
	"testing"

	"github.com/go-openapi/loads"
	"github.com/ve-weiyi/pkg/utils/jsonconv"
)

const SWAGER_PATH = "/Users/weiyi/Github/veweiyi/goctlx/testdata/test.json"

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
