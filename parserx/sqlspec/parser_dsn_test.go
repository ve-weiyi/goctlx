package sqlspec

import (
	"os"
	"testing"
)

// TestParseTables 读取 TEST_MYSQL_DSN 指向的库来解析表结构；未配置时跳过。
func TestParseTables(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过：short 模式不连真实数据库")
	}
	dsn := os.Getenv("TEST_MYSQL_DSN")
	if dsn == "" {
		t.Skip("跳过：未设置 TEST_MYSQL_DSN")
	}

	tables, err := ParseTables(dsn)
	if err != nil {
		t.Fatalf("ParseTables: %v", err)
	}
	if len(tables) == 0 {
		t.Fatal("未解析到任何表")
	}

	for _, table := range tables {
		if table.Name == "" || table.StructName == "" {
			t.Errorf("表元信息不完整: %+v", table)
		}
		if len(table.Fields) == 0 {
			t.Errorf("表 %s 没有字段", table.Name)
		}
	}
}
