package mysql

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ve-weiyi/goctlx/parserx/sqlspec"
)

// TestGenerateModelGolden 用固定 DDL 夹具跑完整链路（解析 → 转换 → 渲染），与 golden 比对。
// 夹具覆盖：跨表同名列类型不同、表级 DEFAULT CHARSET、复合唯一索引的声明列序、普通索引真实名。
func TestGenerateModelGolden(t *testing.T) {
	root := filepath.Join("..", "..", "..")

	tables, err := sqlspec.ParseTablesFromSQL(filepath.Join(root, "testdata", "model", "schema.sql"))
	if err != nil {
		t.Fatalf("ParseTablesFromSQL: %v", err)
	}

	models := make([]*ModelData, 0, len(tables))
	for _, table := range tables {
		models = append(models, ConvertTableMetaToData(table))
	}

	cases := []struct {
		name    string
		tpl     string
		goldens []string
	}{
		{"model", "model.tpl", []string{"t_demo_model.go", "t_demo_log_model.go"}},
		{"thin", "model.thin.tpl", []string{"t_demo_model.go"}},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			out := t.TempDir()
			if err := generateModel(models, filepath.Join(root, "template", "model", c.tpl), out, "%s_model.go"); err != nil {
				t.Fatalf("generateModel: %v", err)
			}
			for _, name := range c.goldens {
				got, err := os.ReadFile(filepath.Join(out, name))
				if err != nil {
					t.Fatalf("read generated %s: %v", name, err)
				}
				wantPath := filepath.Join(root, "testdata", "model", "golden", name+"."+c.name+".txt")
				want, err := os.ReadFile(wantPath)
				if err != nil {
					t.Fatalf("read golden %s: %v", wantPath, err)
				}
				if string(got) != string(want) {
					t.Errorf("%s 与 golden 不一致\n got:\n%s\nwant:\n%s", name, got, want)
				}
			}
		})
	}
}
