package mysql

import (
	"fmt"
	"sort"

	"github.com/zeromicro/go-zero/tools/goctl/util/stringx"

	"github.com/ve-weiyi/goctlx/parserx/sqlspec"
)

func ConvertTableMetaToData(tm *sqlspec.TableMeta) *ModelData {

	var fs []*ModelField
	for _, f := range tm.Fields {
		fs = append(fs, &ModelField{
			Name:    f.Name,
			Type:    f.Type,
			Column:  f.JsonTag,
			Tag:     fmt.Sprintf(`gorm:"%s" json:"%s"`, f.GormTag, f.JsonTag),
			Comment: f.Comment,
		})
	}

	var ufs [][]*ModelField
	keys := make([]string, 0, len(tm.UniqueIndex))
	for k := range tm.UniqueIndex {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		es := tm.UniqueIndex[k]
		u := make([]*ModelField, 0, len(es))
		for _, e := range es {
			u = append(u, &ModelField{
				Name:    e.Name,
				Type:    e.Type,
				Column:  e.JsonTag,
				Tag:     fmt.Sprintf(`gorm:"%s" json:"%s"`, e.GormTag, e.JsonTag),
				Comment: e.Comment,
			})
		}
		ufs = append(ufs, u)
	}

	return &ModelData{
		TableName:           tm.Name,
		UpperStartCamelName: tm.StructName,
		LowerStartCamelName: stringx.From(tm.StructName).Untitle(),
		SnakeName:           tm.Name,
		Fields:              fs,
		UniqueFields:        ufs,
	}
}
