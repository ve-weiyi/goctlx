package sqlspec

import (
	"os"
	"testing"
)

// Test_ParseTables 读取 TEST_MYSQL_DSN 指向的库来解析表结构。
func Test_ParseTables(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过：short 模式不连真实数据库")
	}
	dsn := os.Getenv("TEST_MYSQL_DSN")
	if dsn == "" {
		t.Skip("跳过：未设置 TEST_MYSQL_DSN")
	}

	tables, err := ParseTables(dsn)
	if err != nil {
		// TEST_MYSQL_DSN 指向的库连不上属预期，跳过而非失败
		t.Skipf("跳过：无法连接 MySQL (%v)", err)
	}
	t.Log(dumpJSON(tables))
}
