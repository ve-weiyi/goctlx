package sqlspec

import (
	"fmt"
	"sort"
	"strings"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// ParseTables uses GORM's public API to produce gen-style table metadata.
func ParseTables(dsn string) ([]*TableMeta, error) {
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{SingularTable: true},
	})
	if err != nil {
		return nil, fmt.Errorf("connect db: %w", err)
	}

	tables, err := db.Migrator().GetTables()
	if err != nil {
		return nil, err
	}

	ns := namingStrategy{}

	var result []*TableMeta
	for _, tableName := range tables {
		tm, err := buildTableMeta(db, tableName, ns)
		if err != nil {
			return nil, fmt.Errorf("table %s: %w", tableName, err)
		}
		if tm != nil {
			result = append(result, tm)
		}
	}
	return result, nil
}

func buildTableMeta(db *gorm.DB, tableName string, ns namingStrategy) (*TableMeta, error) {
	columns, err := db.Migrator().ColumnTypes(tableName)
	if err != nil {
		return nil, err
	}

	indexes, err := db.Migrator().GetIndexes(tableName)
	if err != nil {
		return nil, err
	}

	colIndexes := groupIndexes(indexes)

	tm := &TableMeta{
		Name:        tableName,
		StructName:  ns.SchemaName(tableName),
		UniqueIndex: make(map[string][]*FieldMeta),
	}

	for _, col := range columns {
		name := col.Name()
		goType := mapColumnType(col)

		nullable, hasNullable := col.Nullable()
		isPK, _ := col.PrimaryKey()
		goType = wrapPtrType(goType, hasNullable && nullable && !isPK)

		gormTag := buildGormTag(col, goType, colIndexes[name])
		jsonTag := name

		comment, _ := col.Comment()

		fm := &FieldMeta{
			Name:    ns.SchemaName(name),
			Type:    goType,
			GormTag: gormTag,
			JsonTag: jsonTag,
			Comment: comment,
		}
		tm.Fields = append(tm.Fields, fm)

		for _, idx := range colIndexes[name] {
			if idx == nil {
				continue
			}
			if pk, _ := idx.PrimaryKey(); pk {
				continue
			}
			if uniq, _ := idx.Unique(); !uniq {
				continue
			}
			idxName := idx.Name()
			tm.UniqueIndex[idxName] = append(tm.UniqueIndex[idxName], fm)
		}
	}

	for idxName, fields := range tm.UniqueIndex {
		colOrder := make(map[string]int)
		for _, ix := range colIndexes[fields[0].JsonTag] {
			if ix.Name() == idxName {
				for p, c := range ix.Columns() {
					colOrder[c] = p
				}
				break
			}
		}
		sort.SliceStable(fields, func(i, j int) bool {
			return colOrder[fields[i].JsonTag] < colOrder[fields[j].JsonTag]
		})
	}

	return tm, nil
}

func groupIndexes(indexList []gorm.Index) map[string][]gorm.Index {
	result := make(map[string][]gorm.Index)
	for _, idx := range indexList {
		if idx == nil {
			continue
		}
		for _, col := range idx.Columns() {
			result[col] = append(result[col], idx)
		}
	}
	return result
}

func mapColumnType(col gorm.ColumnType) string {
	key := strings.ToLower(col.DatabaseTypeName())

	// if key == "tinyint" {
	// 	colType, _ := col.ColumnType()
	// 	if strings.HasPrefix(strings.TrimSpace(colType), "tinyint(1)") {
	// 		return "bool"
	// 	}
	// }

	typ := mapMySQLType(key)
	//if col.Name() == "deleted_at" && typ == "time.Time" {
	//	return "gorm.DeletedAt"
	//}
	return typ
}

func buildGormTag(col gorm.ColumnType, goType string, indexes []gorm.Index) string {
	colType, _ := col.ColumnType()
	if colType == "" {
		colType = col.DatabaseTypeName()
	}

	parts := []string{
		fmt.Sprintf("column:%s", col.Name()),
		fmt.Sprintf("type:%s", colType),
	}

	isPK, ok := col.PrimaryKey()
	if ok && isPK {
		parts = append(parts, "primaryKey")
		if at, ok := col.AutoIncrement(); ok && at {
			parts = append(parts, "autoIncrement:true")
		}
	} else if n, ok := col.Nullable(); ok && !n {
		parts = append(parts, "not null")
	}

	parts = append(parts, indexParts(col, indexes)...)
	parts = append(parts, defaultParts(col, goType, isPK)...)

	if c, ok := col.Comment(); ok && c != "" {
		parts = append(parts, fmt.Sprintf("comment:%s", c))
	}

	return strings.Join(parts, ";")
}

// indexParts 生成该列参与的所有索引片段，按索引名排序以保证输出稳定。
func indexParts(col gorm.ColumnType, indexes []gorm.Index) []string {
	sort.SliceStable(indexes, func(i, j int) bool {
		if indexes[i] == nil || indexes[j] == nil {
			return false
		}
		ni, nj := indexes[i].Name(), indexes[j].Name()
		if ni == nj {
			return i < j
		}
		return ni < nj
	})

	var parts []string
	for _, idx := range indexes {
		if idx == nil {
			continue
		}
		if pk, _ := idx.PrimaryKey(); pk {
			continue
		}
		prio := 1
		for p, c := range idx.Columns() {
			if c == col.Name() {
				prio = p + 1
				break
			}
		}
		if uniq, _ := idx.Unique(); uniq {
			parts = append(parts, fmt.Sprintf("uniqueIndex:%s,priority:%d", idx.Name(), prio))
		} else {
			parts = append(parts, fmt.Sprintf("index:%s,priority:%d", idx.Name(), prio))
		}
	}
	return parts
}

// defaultParts 生成显式默认值；数据库未显式声明默认值但列 NOT NULL 时补隐式默认值，
// 避免 AutoMigrate 产生 ALTER TABLE 时丢失数据库的隐式默认行为。
func defaultParts(col gorm.ColumnType, goType string, isPK bool) []string {
	var parts []string
	hasDefault := false

	if dv, ok := col.DefaultValue(); ok {
		include := true
		if col.Name() == "created_at" || col.Name() == "updated_at" {
			if goType == "time.Time" {
				trimmed := strings.Trim(dv, "'0:-")
				include = trimmed != ""
			} else {
				include = false
			}
		}
		if include {
			if dv == "" {
				parts = append(parts, "default:''")
			} else {
				parts = append(parts, fmt.Sprintf("default:%s", dv))
			}
			hasDefault = true
		}
	}

	if !hasDefault && !isPK {
		if n, nok := col.Nullable(); nok && !n {
			if col.Name() != "created_at" && col.Name() != "updated_at" {
				baseType := strings.TrimPrefix(goType, "*")
				switch baseType {
				case "string":
					parts = append(parts, "default:''")
				case "int64", "int32", "int", "int16", "int8":
					parts = append(parts, "default:0")
				case "float64", "float32":
					parts = append(parts, "default:0")
				case "bool":
					parts = append(parts, "default:false")
				}
			}
		}
	}
	return parts
}
