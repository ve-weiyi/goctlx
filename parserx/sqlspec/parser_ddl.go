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

// createTableRe 定位每个 CREATE TABLE 的起始位置，用于把整份 DDL 切成单表片段。
var createTableRe = regexp.MustCompile("(?i)CREATE\\s+TABLE\\s+(?:IF\\s+NOT\\s+EXISTS\\s+)?`(\\w+)`")

// splitByTable 把整份 DDL 切成 表名 → 单表片段。
// 列信息必须按表作用域提取：不同表存在同名列（如 created_at 的 datetime 与 datetime(3)），
// 共用一份「列名→值」映射会互相覆盖。
func splitByTable(sql string) map[string]string {
	locs := createTableRe.FindAllStringSubmatchIndex(sql, -1)
	out := make(map[string]string, len(locs))
	for i, loc := range locs {
		end := len(sql)
		if i+1 < len(locs) {
			end = locs[i+1][0]
		}
		out[sql[loc[2]:loc[3]]] = sql[loc[0]:end]
	}
	return out
}

// extractColumnTypes scans SQL DDL text and returns a map of column name → full MySQL type string.
func extractColumnTypes(sql string) map[string]string {
	result := make(map[string]string)
	for _, m := range colTypeRe.FindAllStringSubmatch(sql, -1) {
		result[m[1]] = m[2]
	}
	return result
}

// keyLineRe 匹配表级索引声明行，捕获索引类型前缀、索引名与列清单。
var keyLineRe = regexp.MustCompile("(?mi)^\\s*(PRIMARY\\s+|UNIQUE\\s+|FULLTEXT\\s+|SPATIAL\\s+)?KEY\\s+`(\\w+)`\\s*\\(([^)]*)\\)")

var backtickNameRe = regexp.MustCompile("`(\\w+)`")

// indexDef 是 DDL 里的一条索引声明，columns 保持声明顺序。
type indexDef struct {
	name    string
	columns []string
}

// extractIndexDefs 从单表 DDL 片段里取出唯一索引与普通索引，均按声明顺序返回。
// 索引名与列顺序直接取自 DDL 文本：ddlparser 不保证复合索引的列序，且丢失真实索引名。
func extractIndexDefs(sql string) (unique, plain []indexDef) {
	for _, m := range keyLineRe.FindAllStringSubmatch(sql, -1) {
		kind := strings.ToUpper(strings.TrimSpace(m[1]))
		if kind == "PRIMARY" {
			continue
		}
		names := backtickNameRe.FindAllStringSubmatch(m[3], -1)
		cols := make([]string, 0, len(names))
		for _, n := range names {
			cols = append(cols, n[1])
		}
		if len(cols) == 0 {
			continue
		}
		d := indexDef{name: m[2], columns: cols}
		if kind == "UNIQUE" {
			unique = append(unique, d)
		} else {
			plain = append(plain, d)
		}
	}
	return unique, plain
}

// collectPlainIndexes 把普通索引摊平成「列 → 索引名/序号」，主键列不参与。
// 同一列命中多条索引时取声明顺序的第一条。
func collectPlainIndexes(defs []indexDef, tablePK map[string]bool) (colIndex map[string]string, colIndexPrio map[string]int) {
	colIndex = make(map[string]string)
	colIndexPrio = make(map[string]int)
	for _, d := range defs {
		for i, col := range d.columns {
			if tablePK[col] {
				continue
			}
			if _, exists := colIndex[col]; exists {
				continue
			}
			colIndex[col] = d.name
			colIndexPrio[col] = i + 1
		}
	}
	return colIndex, colIndexPrio
}

// colDefaultRe extracts DEFAULT value from SQL column definitions.
// 限定在同一行内匹配，避免跨行吞掉表选项里的 DEFAULT CHARSET。
var colDefaultRe = regexp.MustCompile("`(\\w+)`[^,\\n]*?DEFAULT\\s+('[^']*'|\\w+(?:\\([^)]*\\))?)")

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
	chunks := splitByTable(sqlText)

	p := ddlparser.NewParser()
	tables, err := p.From(absPath)
	if err != nil {
		return nil, err
	}

	ns := namingStrategy{}
	prefix := filepath.Base(absPath)

	var result []*TableMeta
	for _, t := range tables {
		chunk := chunks[t.Name]
		uniqueNames, plainNames := extractIndexDefs(chunk)
		tm, err := buildTableMetaFromDDL(t, extractColumnTypes(chunk), extractColumnDefaults(chunk), uniqueNames, plainNames, ns)
		if err != nil {
			return nil, fmt.Errorf("%s: table %s: %w", prefix, t.Name, err)
		}
		if tm != nil {
			result = append(result, tm)
		}
	}
	return result, nil
}

func buildTableMetaFromDDL(t *ddlparser.Table, colTypes map[string]string, colDefaults map[string]string, unique, plain []indexDef, ns namingStrategy) (*TableMeta, error) {
	tablePK, colUnique := collectTableConstraints(t, unique)
	colIndex, colIndexPrio := collectPlainIndexes(plain, tablePK)

	tm := &TableMeta{
		Name:        t.Name,
		StructName:  ns.SchemaName(t.Name),
		UniqueIndex: make(map[string][]*FieldMeta),
	}

	for _, col := range t.Columns {
		if col == nil {
			continue
		}

		fm, err := buildFieldFromDDLColumn(col, colTypes, colDefaults, colUnique, colIndex, colIndexPrio, tablePK, ns)
		if err != nil {
			return nil, err
		}
		tm.Fields = append(tm.Fields, fm)

		registerUniqueIndex(tm, fm, col, colUnique, tablePK)
	}

	sortUniqueIndexFields(tm, colUnique)

	return tm, nil
}

// indexRef 是列在某个唯一索引里的位置。
type indexRef struct {
	name string
	prio int
}

// collectTableConstraints 汇总表级约束：主键列集合，以及「列 → 所属唯一索引」。
// 一个列可能同时属于多个唯一索引（如 (a,b) 与 (b,c) 中的 b），因此按列存切片。
func collectTableConstraints(t *ddlparser.Table, unique []indexDef) (tablePK map[string]bool, colUnique map[string][]indexRef) {
	tablePK = make(map[string]bool)         // columns that are table-level primary key
	colUnique = make(map[string][]indexRef) // columnName → 所属唯一索引

	for _, c := range t.Constraints {
		for _, col := range c.ColumnPrimaryKey {
			tablePK[col] = true
		}
	}

	for _, d := range unique {
		for i, col := range d.columns {
			colUnique[col] = append(colUnique[col], indexRef{name: d.name, prio: i + 1})
		}
	}
	return
}

// buildFieldFromDDLColumn 把一列转成 FieldMeta（Go 类型推导 + gorm tag）。
func buildFieldFromDDLColumn(col *ddlparser.Column, colTypes map[string]string, colDefaults map[string]string,
	colUnique map[string][]indexRef, colIndex map[string]string, colIndexPrio map[string]int,
	tablePK map[string]bool, ns namingStrategy) (*FieldMeta, error) {
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

	isPK := (col.Constraint != nil && col.Constraint.Primary) || tablePK[col.Name]
	isNullable := !isPK && (col.Constraint == nil || !col.Constraint.NotNull)
	goType = wrapPtrType(goType, isNullable)

	fm := &FieldMeta{
		Name:    ns.SchemaName(col.Name),
		Type:    goType,
		GormTag: buildDDLGormTag(col, fullType, colUnique, colIndex, colIndexPrio, tablePK, colDefaults),
		JsonTag: col.Name,
	}

	if col.Constraint != nil {
		fm.Comment = col.Constraint.Comment
	}
	return fm, nil
}

// registerUniqueIndex 把该列登记进它所属的每个唯一索引（主键列不参与）。
func registerUniqueIndex(tm *TableMeta, fm *FieldMeta, col *ddlparser.Column, colUnique map[string][]indexRef, tablePK map[string]bool) {
	if tablePK[col.Name] {
		return
	}
	if col.Constraint != nil && col.Constraint.Primary {
		return
	}

	for _, ref := range colUnique[col.Name] {
		fmCopy := *fm
		tm.UniqueIndex[ref.name] = append(tm.UniqueIndex[ref.name], &fmCopy)
	}

	if col.Constraint != nil && col.Constraint.Unique {
		idxName := "uk_" + col.Name
		if !hasIndexRef(colUnique[col.Name], idxName) {
			fmCopy := *fm
			tm.UniqueIndex[idxName] = append(tm.UniqueIndex[idxName], &fmCopy)
		}
	}
}

// sortUniqueIndexFields 让每个唯一索引内的字段按索引声明顺序排列。
// 列是按表字段顺序登记的，跨列索引必须按索引内的位置重排，方法名才与声明一致。
func sortUniqueIndexFields(tm *TableMeta, colUnique map[string][]indexRef) {
	for idxName, fields := range tm.UniqueIndex {
		sort.SliceStable(fields, func(i, j int) bool {
			return indexPrio(colUnique, fields[i].JsonTag, idxName) < indexPrio(colUnique, fields[j].JsonTag, idxName)
		})
	}
}

func indexPrio(colUnique map[string][]indexRef, col, idxName string) int {
	for _, r := range colUnique[col] {
		if r.name == idxName {
			return r.prio
		}
	}
	return 0
}

func hasIndexRef(refs []indexRef, name string) bool {
	for _, r := range refs {
		if r.name == name {
			return true
		}
	}
	return false
}

func buildDDLGormTag(col *ddlparser.Column, mysqlType string, colUnique map[string][]indexRef,
	colIndex map[string]string, colIndexPrio map[string]int, tablePK map[string]bool, colDefaults map[string]string) string {
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
		for _, ref := range colUnique[col.Name] {
			parts = append(parts, fmt.Sprintf("uniqueIndex:%s,priority:%d", ref.name, ref.prio))
		}
		if c.Unique {
			parts = append(parts, fmt.Sprintf("uniqueIndex:uk_%s,priority:1", col.Name))
		}
		if idxName, ok := colIndex[col.Name]; ok {
			parts = append(parts, fmt.Sprintf("index:%s,priority:%d", idxName, colIndexPrio[col.Name]))
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
