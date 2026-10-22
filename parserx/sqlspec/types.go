package sqlspec

import (
	"github.com/zeromicro/go-zero/tools/goctl/util/stringx"

	"gorm.io/gorm/schema"
)

// namingStrategy converts snake_case to CamelCase without treating common
// initialisms specially. "user_id" → "UserId".
type namingStrategy struct {
	schema.NamingStrategy
}

func (namingStrategy) SchemaName(table string) string {
	return toCamelCase(table)
}

// TableMeta holds model info produced from database table metadata
type TableMeta struct {
	Name        string // "t_user"
	StructName  string // "TUser"
	Comment     string // struct comment
	Fields      []*FieldMeta
	UniqueIndex map[string][]*FieldMeta // indexName → fields
}

// FieldMeta describes a struct field
type FieldMeta struct {
	Name    string // "ID"
	Type    string // "int64"
	GormTag string // "column:id;type:bigint;primaryKey;autoIncrement:true;comment:ID"
	JsonTag string // "id"
	Comment string // field line comment
}

// dataTypeMap maps MySQL type names to Go types, replicating gorm.io/gen's default mapping.
var dataTypeMap = map[string]func(string) string{
	"tinyint":   func(string) string { return "int64" },
	"smallint":  func(string) string { return "int64" },
	"mediumint": func(string) string { return "int64" },
	"bigint":    func(string) string { return "int64" },
	"int":       func(string) string { return "int64" },

	"numeric":    func(string) string { return "int32" },
	"integer":    func(string) string { return "int32" },
	"float":      func(string) string { return "float32" },
	"real":       func(string) string { return "float64" },
	"double":     func(string) string { return "float64" },
	"decimal":    func(string) string { return "float64" },
	"char":       func(string) string { return "string" },
	"varchar":    func(string) string { return "string" },
	"tinytext":   func(string) string { return "string" },
	"mediumtext": func(string) string { return "string" },
	"longtext":   func(string) string { return "string" },
	"binary":     func(string) string { return "[]byte" },
	"varbinary":  func(string) string { return "[]byte" },
	"tinyblob":   func(string) string { return "[]byte" },
	"blob":       func(string) string { return "[]byte" },
	"mediumblob": func(string) string { return "[]byte" },
	"longblob":   func(string) string { return "[]byte" },
	"text":       func(string) string { return "string" },
	"json":       func(string) string { return "string" },
	"enum":       func(string) string { return "string" },
	"time":       func(string) string { return "time.Time" },
	"date":       func(string) string { return "time.Time" },
	"datetime":   func(string) string { return "time.Time" },
	"timestamp":  func(string) string { return "time.Time" },
	"year":       func(string) string { return "int32" },
	"bit":        func(string) string { return "[]byte" },
	"boolean":    func(string) string { return "bool" },
}

// toCamelCase converts snake_case to CamelCase（复用 goctl 的 stringx，
// 与本仓 cmd/api/gin 用的实现保持同一套，避免同仓两份驼峰转换）。
func toCamelCase(s string) string {
	return stringx.From(s).ToCamel()
}

// mapMySQLType maps a MySQL type name to its Go type.
func mapMySQLType(mysqlType string) string {
	if fn, ok := dataTypeMap[mysqlType]; ok {
		return fn("")
	}
	return "string"
}

// wrapPtrType converts a nullable Go type to its sql.Null* equivalent.
func wrapPtrType(goType string, nullable bool) string {
	if !nullable || goType == "gorm.DeletedAt" {
		return goType
	}
	return "*" + goType
	//switch goType {
	//case "string":
	//	return "sql.NullString"
	//case "int64":
	//	return "sql.NullInt64"
	//case "int32":
	//	return "sql.NullInt32"
	//case "float64":
	//	return "sql.NullFloat64"
	//case "bool":
	//	return "sql.NullBool"
	//case "time.Time":
	//	return "sql.NullTime"
	//}
	//return goType
}
