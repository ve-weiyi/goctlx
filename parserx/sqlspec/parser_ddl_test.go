package sqlspec

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
)

func Test_ParseTablesFromSQL(t *testing.T) {
	const sqlFile = "testdata/test.sql"
	if _, err := os.Stat(sqlFile); err != nil {
		t.Skipf("跳过：缺少 SQL 夹具 %s", sqlFile)
	}

	tables, err := ParseTablesFromSQL(sqlFile)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(dumpJSON(tables))
}

// dumpJSON 把值序列化为缩进 JSON，便于测试输出查看
func dumpJSON(v any) string {
	data, err := json.MarshalIndent(v, "", " ")
	if err != nil || string(data) == "{}" {
		return fmt.Sprintf("%+v", v)
	}
	return string(data)
}
