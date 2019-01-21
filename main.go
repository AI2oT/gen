package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"go/format"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/AI2oT/gen/dbmeta"
	gtmpl "github.com/AI2oT/gen/template"
	"github.com/droundy/goopt"
	"github.com/jimsmart/schema"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/jinzhu/inflection"
	_ "github.com/lib/pq"
	"github.com/serenize/snaker"
)

var (
	sqlConnStr  = goopt.String([]string{"-c", "--connstr"}, "nil", "database connection string")
	sqlDatabase = goopt.String([]string{"-d", "--database"}, "nil", "Database to for connection")
	sqlTable    = goopt.String([]string{"-t", "--table"}, "", "Table to build struct from")

	packageName = goopt.String([]string{"--package"}, "", "name to set for package")

	jsonAnnotation = goopt.Flag([]string{"--json"}, []string{"--no-json"}, "Add json annotations (default)", "Disable json annotations")
	gormAnnotation = goopt.Flag([]string{"--gorm"}, []string{}, "Add gorm annotations (tags)", "")

	verbose = goopt.Flag([]string{"-v", "--verbose"}, []string{}, "Enable verbose output", "")
)

func init() {
	// Setup goopts
	goopt.Description = func() string {
		return "ORM generator for Mysql"
	}
	goopt.Version = "0.1"
	goopt.Summary = `gen [-v] --connstr "user:password@/dbname" --package pkgName --database databaseName --table tableName [--json] [--gorm] [--guregu]`

	//Parse options
	goopt.Parse(nil)

}

func main() {
	// Username is required
	if sqlConnStr == nil || *sqlConnStr == "" {
		fmt.Println("sql connection string is required! Add it with --connstr=s")
		return
	}

	if sqlDatabase == nil || *sqlDatabase == "" {
		fmt.Println("Database can not be null")
		return
	}

	var db, err = sql.Open("mysql", *sqlConnStr)
	if err != nil {
		fmt.Println("Error in open database: " + err.Error())
		return
	}
	defer db.Close()

	// parse or read tables
	var tables []string
	if *sqlTable != "" {
		tables = strings.Split(*sqlTable, ",")
	} else {
		tables, err = schema.TableNames(db)
		if err != nil {
			fmt.Println("Error in fetching tables information from mysql information schema")
			return
		}
	}
	// if packageName is not set we need to default it
	if packageName == nil || *packageName == "" {
		*packageName = "generated"
	}
	os.Mkdir("model", 0777)

	t, err := getTemplate(gtmpl.ModelTmpl)
	if err != nil {
		fmt.Println("Error in loading model template: " + err.Error())
		return
	}

	var structNames []string

	// generate go files for each table
	for _, tableName := range tables {
		structName := dbmeta.FmtFieldName(tableName)
		structName = inflection.Singular(structName)
		structNames = append(structNames, structName)

		modelInfo := dbmeta.GenerateStruct(db, tableName, structName, "model", *jsonAnnotation, *gormAnnotation)

		var buf bytes.Buffer
		err = t.Execute(&buf, modelInfo)
		if err != nil {
			fmt.Println("Error in rendering model: " + err.Error())
			return
		}
		data, err := format.Source(buf.Bytes())
		if err != nil {
			fmt.Println("Error in formating source: " + err.Error())
			return
		}
		ioutil.WriteFile(filepath.Join("model", inflection.Singular(tableName)+".go"), data, 0777)

	}

}

func getTemplate(t string) (*template.Template, error) {
	var funcMap = template.FuncMap{
		"pluralize":        inflection.Plural,
		"title":            strings.Title,
		"toLower":          strings.ToLower,
		"toLowerCamelCase": camelToLowerCamel,
		"toSnakeCase":      snaker.CamelToSnake,
	}

	tmpl, err := template.New("model").Funcs(funcMap).Parse(t)

	if err != nil {
		return nil, err
	}

	return tmpl, nil
}

func camelToLowerCamel(s string) string {
	ss := strings.Split(s, "")
	ss[0] = strings.ToLower(ss[0])

	return strings.Join(ss, "")
}
