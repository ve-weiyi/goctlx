package mysql

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/zeromicro/go-zero/tools/goctl/util"

	"github.com/ve-weiyi/goctlx/x/gofile"
)

type (
	ModelData struct {
		TableName           string
		UpperStartCamelName string
		LowerStartCamelName string
		SnakeName           string
		Fields              []*ModelField
		UniqueFields        [][]*ModelField
	}

	ModelField struct {
		Name    string // 属性名称  Name
		Type    string // 属性类型  string、int、bool、float、{UpperStartCamelName}
		Column  string // 数据库列名
		Tag     string // json tag
		Comment string // 行尾注释
	}
)

func generateModel(models []*ModelData, tplPath string, outPath string, nameAs string) error {
	tpl, err := os.ReadFile(tplPath)
	if err != nil {
		return err
	}

	t := util.With("model").Parse(string(tpl))
	for name, fn := range map[string]any{
		"funcFieldsKey": func(fs []*ModelField) string {
			var name string
			for _, ff := range fs {
				name += ff.Name
			}
			return name
		},
		"funcFieldsKeyVar": func(fs []*ModelField) string {
			var name string
			for _, ff := range fs {
				v := ff.Column
				tp := ff.Type
				if name != "" {
					name += ", "
				}
				name += fmt.Sprintf("%s %s", v, extractBaseType(tp))
			}
			return name
		},
		"funcFieldsKeyCond": func(fs []*ModelField) string {
			var name string
			for _, ff := range fs {
				v := ff.Column
				if name != "" {
					name += " and "
				}
				name += fmt.Sprintf("`%s` = ?", v)
			}
			return name
		},
		"funcFieldsKeyCondVar": func(fs []*ModelField) string {
			var name string
			for _, ff := range fs {
				v := ff.Column
				if name != "" {
					name += ", "
				}
				name += v
			}
			return name
		},
	} {
		t.AddFunc(name, fn)
	}

	for _, model := range models {
		buf, err := t.Execute(model)
		if err != nil {
			return err
		}
		if err := gofile.Write(path.Join(outPath, fmt.Sprintf(nameAs, model.TableName)), buf.Bytes()); err != nil {
			return err
		}
	}

	return nil
}

// extractBaseType 提取基础类型原型
// sql.NullString -> string
// sql.NullInt64 -> int64
// sql.NullTime -> time.Time
// string -> string (保持不变)
func extractBaseType(dataType string) string {
	switch dataType {
	case "sql.NullString":
		return "string"
	case "sql.NullInt64":
		return "int64"
	case "sql.NullInt32":
		return "int32"
	case "sql.NullFloat64":
		return "float64"
	case "sql.NullBool":
		return "bool"
	case "sql.NullTime":
		return "time.Time"
	default:
		return strings.TrimPrefix(dataType, "*")
	}
}
