/**
 * @package gm_cli
 * @file      : main.go
 * @author    : LeiXiaoTian
 * @contact   : 1124378213@qq.com
 * @time      : 2024/5/15 16:14
 **/
package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"gopkg.in/ini.v1"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Config struct {
	Link         string
	Group        string
	Prefix       string
	RemovePrefix string
	JsonCase     string
	Debug        bool
	Tables       string
	GenDIr       string
}

type Column struct {
	ColumnName    string `gorm:"column:COLUMN_NAME"`
	ColumnType    string `gorm:"column:COLUMN_TYPE"`
	IsNullable    string `gorm:"column:IS_NULLABLE"`
	ColumnKey     string `gorm:"column:COLUMN_KEY"`
	ColumnComment string `gorm:"column:COLUMN_COMMENT"`
}


func loadConfig() (*Config, error) {
	cfg, err := ini.Load("config.ini")
	if err != nil {
		return nil, err
	}

	section := cfg.Section("mysql")
	return &Config{
		Link:         section.Key("link").String(),
		Group:        section.Key("group").String(),
		Prefix:       section.Key("prefix").String(),
		RemovePrefix: section.Key("removePrefix").String(),
		JsonCase:     section.Key("jsonCase").String(),
		Debug:        section.Key("debug").MustBool(false),
		Tables:       section.Key("tables").String(),
		GenDIr:       section.Key("genDir").String(),
	}, nil
}

func toCamelCase(s string) string {
	parts := strings.Split(s, "_")
	for i := range parts {
		parts[i] = strings.Title(parts[i])
	}
	return strings.Join(parts, "")
}

func mapColumnType(mysqlType string) string {
	switch {
	case strings.HasPrefix(mysqlType, "int"):
		return "int"
	case strings.HasPrefix(mysqlType, "bigint"):
		return "int64"
	case strings.HasPrefix(mysqlType, "varchar"), strings.HasPrefix(mysqlType, "text"):
		return "string"
	case strings.HasPrefix(mysqlType, "datetime"), strings.HasPrefix(mysqlType, "timestamp"):
		return "string"
	case strings.HasPrefix(mysqlType, "decimal"), strings.HasPrefix(mysqlType, "float"), strings.HasPrefix(mysqlType, "double"):
		return "float64"
	default:
		return "string"
	}
}


func fetchColumns(db *gorm.DB, tableName string) ([]Column, error) {
	var columns []Column
	query := fmt.Sprintf("SELECT COLUMN_NAME, COLUMN_TYPE, IS_NULLABLE, COLUMN_KEY, COLUMN_COMMENT FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = '%s'", tableName)
	if err := db.Raw(query).Scan(&columns).Error; err != nil {
		return nil, err
	}
	return columns, nil
}

func generateFile(tableName, outDir, prefix, removePrefix string, columns []Column) error {
	structName := toCamelCase(tableName)

	// Apply removePrefix if it's not empty
	if removePrefix != "" && strings.HasPrefix(tableName, removePrefix) {
		tableName = strings.TrimPrefix(tableName, removePrefix)
	}

	// Apply prefix if it's not empty
	if prefix != "" {
		tableName = prefix + tableName
	}

	// Ensure output directory exists
	if err := os.MkdirAll(outDir, os.ModePerm); err != nil {
		return err
	}

	// Create the output file
	filePath := fmt.Sprintf("%s/%s.go", outDir, tableName)
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Create the struct fields from columns
	var fields strings.Builder
	for _, col := range columns {
		fieldName := toCamelCase(col.ColumnName)
		fieldType := mapColumnType(col.ColumnType)
		jsonTag := strings.ToLower(col.ColumnName)
		comment := col.ColumnComment
		if comment == "" {
			comment = fieldName
		}
		fields.WriteString(fmt.Sprintf("    %s %s `json:\"%s\"` // %s\n", fieldName, fieldType, jsonTag, comment))
	}

	// Define the template for the struct
	structTemplate := `package {{.PackageName}}

type {{.StructName}} struct {
{{.Fields}}
}

// 返回表名
func (this {{.StructName}}) TableName() string {
	return "{{.TableName}}"
}
`
	// 获取最后一级文件夹名
	packageName := filepath.Base(outDir)

	// Parse the template
	tmpl, err := template.New("goStruct").Parse(structTemplate)
	if err != nil {
		return err
	}

	// Execute the template with the table name and fields
	data := struct {
		PackageName string
		StructName  string
		TableName   string
		Fields      string
	}{
		PackageName: packageName,
		StructName: structName,
		TableName:  tableName,
		Fields:     fields.String(),
	}

	if err := tmpl.Execute(file, data); err != nil {
		return err
	}

	// Format the file with gofmt
	cmd := exec.Command("gofmt", "-w", filePath)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to format file with gofmt: %v", err)
	}

	return nil
}


func main() {
	config, err := loadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	fmt.Println("link:",config.Link)
	// Connect to MySQL
	db, err := gorm.Open(mysql.Open(config.Link), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	tableNames := strings.Split(config.Tables, ",")

	for _, tableName := range tableNames {
		// Fetch columns for the table
		columns, err := fetchColumns(db, tableName)
		if err != nil {
			log.Fatalf("Failed to fetch columns for table %s: %v", tableName, err)
		}

		// Generate file for the table
		err = generateFile(tableName, config.GenDIr, config.Prefix, config.RemovePrefix, columns)
		if err != nil {
			log.Fatalf("Failed to generate file for table %s: %v", tableName, err)
		}
		fmt.Printf("Successfully generated file for table %s\n", tableName)
	}
}


