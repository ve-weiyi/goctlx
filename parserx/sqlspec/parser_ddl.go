package sqlspec

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	ddlparser "github.com/zeromicro/ddl-parser/parser"
)

// ddlTypeToMySQL maps ddl-parser data type enum constants to MySQL type name strings.
var ddlTypeToMySQL = map[int]string{
	ddlparser.Bit:                "bit",
	ddlparser.TinyInt:            "tinyint",
	ddlparser.SmallInt:           "smallint",
	ddlparser.MediumInt:          "mediumint",
	ddlparser.Int:                "int",
	ddlparser.Integer:            "integer",
	ddlparser.BigInt:             "bigint",
	ddlparser.MiddleInt:          "middleint",
	ddlparser.Int1:               "int1",
	ddlparser.Int2:               "int2",
	ddlparser.Int3:               "int3",
	ddlparser.Int4:               "int4",
	ddlparser.Int8:               "int8",
	ddlparser.Float:              "float",
	ddlparser.Float4:             "float4",
	ddlparser.Float8:             "float8",
	ddlparser.Double:             "double",
	ddlparser.Decimal:            "decimal",
	ddlparser.Dec:                "dec",
	ddlparser.Fixed:              "fixed",
	ddlparser.Numeric:            "numeric",
	ddlparser.Real:               "real",
	ddlparser.Bool:               "bool",
	ddlparser.Boolean:            "boolean",
	ddlparser.Date:               "date",
	ddlparser.DateTime:           "datetime",
	ddlparser.Timestamp:          "timestamp",
	ddlparser.Time:               "time",
	ddlparser.Year:               "year",
	ddlparser.Char:               "char",
	ddlparser.Character:          "character",
	ddlparser.VarChar:            "varchar",
	ddlparser.NChar:              "nchar",
	ddlparser.NVarChar:           "nvarchar",
	ddlparser.TinyText:           "tinytext",
	ddlparser.Text:               "text",
	ddlparser.MediumText:         "mediumtext",
	ddlparser.LongText:           "longtext",
	ddlparser.Binary:             "binary",
	ddlparser.VarBinary:          "varbinary",
	ddlparser.TinyBlob:           "tinyblob",
	ddlparser.Blob:               "blob",
	ddlparser.MediumBlob:         "mediumblob",
	ddlparser.LongBlob:           "longblob",
	ddlparser.Enum:               "enum",
	ddlparser.Set:                "set",
	ddlparser.Json:               "json",
	ddlparser.LongVarChar:        "longvarchar",
	ddlparser.LongVarBinary:      "longvarbinary",
	ddlparser.GeometryCollection: "geometrycollection",
	ddlparser.GeomCollection:     "geomcollection",
	ddlparser.LineString:         "linestring",
	ddlparser.MultiLineString:    "multilinestring",
	ddlparser.MultiPoint:         "multipoint",
	ddlparser.MultiPolygon:       "multipolygon",
	ddlparser.Point:              "point",
	ddlparser.Polygon:            "polygon",
	ddlparser.Geometry:           "geometry",
	ddlparser.Serial:             "serial",
}

var colTypeRe = regexp.MustCompile("`(\\w+)`\\s+(\\w+(?:\\([^)]*\\))?(?:\\s+unsigned)?(?:\\s+zerofill)?)")

// extractColumnTypes scans SQL DDL text and returns a map of column name → full MySQL type string.
func extractColumnTypes(sql string) map[string]string {
	result := make(map[string]string)
	for _, m := range colTypeRe.FindAllStringSubmatch(sql, -1) {
		result[m[1]] = m[2]
	}
	return result
}

// colDefaultRe extracts DEFAULT value from SQL column definitions.
var colDefaultRe = regexp.MustCompile("`(\\w+)`[^,]*?DEFAULT\\s+('[^']*'?|[^\\s,)]+)")

func extractColumnDefaults(sql string) map[string]string {
	result := make(map[string]string)
	for _, m := range colDefaultRe.FindAllStringSubmatch(sql, -1) {
		dv := m[2]
		dv = strings.Trim(dv, "'")
		if strings.EqualFold(dv, "NULL") {
			continue // DEFAULT NULL 不算显式默认值
		}
		result[m[1]] = dv
	}
	return result
}

// ParseTablesFromSQL parses a SQL DDL file and returns gen-style TableMeta.
func ParseTablesFromSQL(filename string) ([]*TableMeta, error) {
	absPath, err := filepath.Abs(filename)
	if err != nil {
		return nil, fmt.Errorf("resolve path %s: %w", filename, err)
	}

	sqlBytes, err := os.ReadFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("read file %s: %w", absPath, err)
	}
	sqlText := string(sqlBytes)
	colTypes := extractColumnTypes(sqlText)
	colDefaults := extractColumnDefaults(sqlText)

	p := ddlparser.NewParser()
	tables, err := p.From(absPath)
	if err != nil {
		return nil, err
	}

	ns := namingStrategy{}
	prefix := filepath.Base(absPath)

	var result []*TableMeta
	for _, t := range tables {
		tm, err := buildTableMetaFromDDL(t, colTypes, colDefaults, ns)
		if err != nil {
			return nil, fmt.Errorf("%s: table %s: %w", prefix, t.Name, err)
		}
		if tm != nil {
			result = append(result, tm)
		}
	}
	return result, nil
}

func buildTableMetaFromDDL(t *ddlparser.Table, colTypes map[string]string, colDefaults map[string]string, ns namingStrategy) (*TableMeta, error) {
	multiColUnique := make(map[string]string) // columnName → indexName
	multiColPrio := make(map[string]int)      // columnName → priority
	tablePK := make(map[string]bool)          // columns that are table-level primary key

	for _, c := range t.Constraints {
		if len(c.ColumnPrimaryKey) > 0 {
			for _, col := range c.ColumnPrimaryKey {
				tablePK[col] = true
			}
		}
		if len(c.ColumnUniqueKey) > 0 {
			idxName := "uk_" + strings.Join(c.ColumnUniqueKey, "_")
			for i, col := range c.ColumnUniqueKey {
				multiColUnique[col] = idxName
				multiColPrio[col] = i + 1
			}
		}
	}

	tm := &TableMeta{
		Name:        t.Name,
		StructName:  ns.SchemaName(t.Name),
		UniqueIndex: make(map[string][]*FieldMeta),
	}

	for _, col := range t.Columns {
		if col == nil {
			continue
		}

		baseType, ok := ddlTypeToMySQL[col.DataType.Type()]
		if !ok {
			return nil, fmt.Errorf("unsupported data type enum: %v", col.DataType.Type())
		}

		// Use full type from SQL text if available, otherwise fall back to base type
		fullType := baseType
		if ft, ok := colTypes[col.Name]; ok {
			fullType = ft
		}

		goType := mapMySQLType(baseType)

		if col.Name == "deleted_at" && goType == "time.Time" {
			goType = "gorm.DeletedAt"
		}

		isPK := (col.Constraint != nil && col.Constraint.Primary) || tablePK[col.Name]
		isNullable := !isPK && (col.Constraint == nil || !col.Constraint.NotNull)
		goType = wrapPtrType(goType, isNullable)

		gormTag := buildDDLGormTag(col, fullType, multiColUnique, multiColPrio, tablePK, colDefaults)

		fm := &FieldMeta{
			Name:    ns.SchemaName(col.Name),
			Type:    goType,
			GormTag: gormTag,
			JsonTag: col.Name,
		}

		if col.Constraint != nil {
			fm.Comment = col.Constraint.Comment
		}

		tm.Fields = append(tm.Fields, fm)

		if col.Constraint != nil && !(col.Constraint.Primary || tablePK[col.Name]) {
			if idxName, ok := multiColUnique[col.Name]; ok {
				fmCopy := *fm
				tm.UniqueIndex[idxName] = append(tm.UniqueIndex[idxName], &fmCopy)
			}
			if col.Constraint.Unique {
				idxName := "uk_" + col.Name
				if _, exists := multiColUnique[col.Name]; !exists || idxName != multiColUnique[col.Name] {
					fmCopy := *fm
					tm.UniqueIndex[idxName] = append(tm.UniqueIndex[idxName], &fmCopy)
				}
			}
		}
	}

	for _, fields := range tm.UniqueIndex {
		sort.SliceStable(fields, func(i, j int) bool {
			return multiColPrio[fields[i].JsonTag] < multiColPrio[fields[j].JsonTag]
		})
	}

	return tm, nil
}

func buildDDLGormTag(col *ddlparser.Column, mysqlType string, multiColUnique map[string]string, multiColPrio map[string]int, tablePK map[string]bool, colDefaults map[string]string) string {
	parts := []string{
		fmt.Sprintf("column:%s", col.Name),
		fmt.Sprintf("type:%s", mysqlType),
	}

	if col.Constraint == nil {
		return strings.Join(parts, ";")
	}

	c := col.Constraint
	isPK := c.Primary || tablePK[col.Name]

	if isPK {
		parts = append(parts, "primaryKey")
		if c.AutoIncrement {
			parts = append(parts, "autoIncrement:true")
		}
	} else if c.NotNull {
		parts = append(parts, "not null")
	}

	if !isPK {
		if idxName, ok := multiColUnique[col.Name]; ok {
			prio := multiColPrio[col.Name]
			parts = append(parts, fmt.Sprintf("uniqueIndex:%s,priority:%d", idxName, prio))
		}
		if c.Unique {
			parts = append(parts, fmt.Sprintf("uniqueIndex:uk_%s,priority:1", col.Name))
		} else if c.Key {
			parts = append(parts, fmt.Sprintf("index:idx_%s,priority:1", col.Name))
		}
	}

	hasDefault := false
	if dv, ok := colDefaults[col.Name]; ok {
		if col.Name == "created_at" || col.Name == "updated_at" {
			// 时间列有 CURRENT_TIMESTAMP 等默认值才写入
			trimmed := strings.Trim(dv, "'0:-")
			if trimmed != "" {
				parts = append(parts, fmt.Sprintf("default:%s", dv))
				hasDefault = true
			}
		} else {
			if dv == "" {
				parts = append(parts, "default:''")
			} else {
				parts = append(parts, fmt.Sprintf("default:%s", dv))
			}
			hasDefault = true
		}
	}

	// 如果 DDL 中没有显式默认值但列是 NOT NULL，补充隐式默认值
	if !hasDefault && !isPK && c.NotNull {
		if col.Name != "created_at" && col.Name != "updated_at" {
			baseType := extractMySQLBaseType(mysqlType)
			switch baseType {
			case "varchar", "char", "text", "tinytext", "mediumtext", "longtext", "enum", "json":
				parts = append(parts, "default:''")
			case "int", "bigint", "tinyint", "smallint", "mediumint", "integer", "decimal", "numeric", "float", "double", "real":
				parts = append(parts, "default:0")
			case "bool", "boolean":
				parts = append(parts, "default:false")
			}
		}
	}

	if c.Comment != "" {
		parts = append(parts, fmt.Sprintf("comment:%s", c.Comment))
	}

	return strings.Join(parts, ";")
}

func extractMySQLBaseType(mysqlType string) string {
	// 从 "varchar(64)" 或 "decimal(10,2)" 中提取基础类型名
	mysqlType = strings.ToLower(strings.TrimSpace(mysqlType))
	if idx := strings.IndexByte(mysqlType, '('); idx >= 0 {
		return mysqlType[:idx]
	}
	return mysqlType
}
