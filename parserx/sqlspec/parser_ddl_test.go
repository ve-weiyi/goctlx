package sqlspec

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseTablesFromSQL(t *testing.T) {
	const schema = "CREATE TABLE `t_user` (\n" +
		"  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键',\n" +
		"  `nickname` varchar(64) NOT NULL DEFAULT '' COMMENT '昵称',\n" +
		"  PRIMARY KEY (`id`)\n" +
		") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户';"

	path := filepath.Join(t.TempDir(), "schema.sql")
	if err := os.WriteFile(path, []byte(schema), 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	tables, err := ParseTablesFromSQL(path)
	if err != nil {
		t.Fatalf("ParseTablesFromSQL: %v", err)
	}
	if len(tables) != 1 {
		t.Fatalf("表数量 = %d, want 1", len(tables))
	}

	table := tables[0]
	if table.Name != "t_user" {
		t.Errorf("表名 = %q, want t_user", table.Name)
	}
	if len(table.Fields) != 2 {
		t.Fatalf("字段数 = %d, want 2", len(table.Fields))
	}
	if table.Fields[0].JsonTag != "id" || !strings.Contains(table.Fields[0].GormTag, "primaryKey") {
		t.Errorf("主键字段 = %+v", table.Fields[0])
	}
	if table.Fields[1].JsonTag != "nickname" {
		t.Errorf("第二字段 = %+v", table.Fields[1])
	}
}
