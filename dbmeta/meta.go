package dbmeta

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/jimsmart/schema"
)

type ModelInfo struct {
	PackageName     string
	StructName      string
	ShortStructName string
	TableName       string
	Fields          []string
}

// commonInitialisms is a set of common initialisms.
// Only add entries that are highly unlikely to be non-initialisms.
// For instance, "ID" is fine (Freudian code is rare), but "AND" is not.
var commonInitialisms = map[string]bool{
	"API":   true,
	"ASCII": true,
	"CPU":   true,
	"CSS":   true,
	"DNS":   true,
	"EOF":   true,
	"GUID":  true,
	"HTML":  true,
	"HTTP":  true,
	"HTTPS": true,
	"ID":    true,
	"IP":    true,
	"JSON":  true,
	"LHS":   true,
	"QPS":   true,
	"RAM":   true,
	"RHS":   true,
	"RPC":   true,
	"SLA":   true,
	"SMTP":  true,
	"SSH":   true,
	"TLS":   true,
	"TTL":   true,
	"UI":    true,
	"UID":   true,
	"UUID":  true,
	"URI":   true,
	"URL":   true,
	"UTF8":  true,
	"VM":    true,
	"XML":   true,
}

var intToWordMap = []string{
	"zero",
	"one",
	"two",
	"three",
	"four",
	"five",
	"six",
	"seven",
	"eight",
	"nine",
}

// GenerateStruct generates a struct for the given table.
func GenerateStruct(db *sql.DB, tableName string, structName string, pkgName string, jsonAnnotation bool, gormAnnotation bool) *ModelInfo {
	cols, _ := schema.Table(db, tableName)
	fields := generateFieldsTypes(db, cols, 0, jsonAnnotation, gormAnnotation)

	var modelInfo = &ModelInfo{
		PackageName:     pkgName,
		StructName:      structName,
		TableName:       tableName,
		ShortStructName: strings.ToLower(string(structName[0])),
		Fields:          fields,
	}

	return modelInfo
}

// Generate fields string
func generateFieldsTypes(db *sql.DB, columns []*sql.ColumnType, depth int, jsonAnnotation bool, gormAnnotation bool) []string {

	//sort.Strings(keys)

	var fields []string
	var field = ""
	for i, c := range columns {
		nullable, _ := c.Nullable()
		key := c.Name()
		valueType := sqlTypeToGoType(strings.ToLower(c.DatabaseTypeName()), nullable)
		if valueType == "" { // unknown type
			continue
		}
		fieldName := FmtFieldName(stringifyFirstChar(key))

		var annotations []string
		if gormAnnotation == true {
			if i == 0 {
				annotations = append(annotations, fmt.Sprintf("gorm:\"column:%s;primary_key\"", key))
			} else {
				annotations = append(annotations, fmt.Sprintf("gorm:\"column:%s\"", key))
			}

		}
		if jsonAnnotation == true {
			annotations = append(annotations, fmt.Sprintf("json:\"%s\"", key))
		}
		if len(annotations) > 0 {
			field = fmt.Sprintf("%s %s `%s`",
				fieldName,
				valueType,
				strings.Join(annotations, " "))

		} else {
			field = fmt.Sprintf("%s %s",
				fieldName,
				valueType)
		}

		fields = append(fields, field)
	}
	return fields
}

func sqlTypeToGoType(mysqlType string, nullable bool) string {
	switch mysqlType {
	case "tinyint":
		if nullable {
			return "*int8"
		}
		return "int8"

	case "int":
		if nullable {
			return "*int32"
		}
		return "int32"
	case "bigint":
		if nullable {
			return "*int64"
		}
		return "int64"
	case "char", "enum", "varchar", "longtext", "mediumtext", "text", "tinytext":
		if nullable {
			return "*string"
		}
		return "string"
	case "date", "datetime", "time", "timestamp":
		if nullable {
			return "*time.Time"
		}
		return "time.Time"
	case "decimal", "double":
		if nullable {
			return "*float64"
		}
		return "float64"
	case "float":
		if nullable {
			return "*float32"
		}
		return "float32"
	case "binary", "blob", "longblob", "mediumblob", "varbinary":
		return "[]byte"
	default:
		panic("unsupported mysql type " + mysqlType)
	}
	return ""
}
