package db

import (
	"fmt"
	"strings"

	"github.com/go-errors/errors"
	"gorm.io/gorm"
)

type Result struct {
	Name string
}

func Version(db *gorm.DB) (string, error) {
	var version string
	tx := db.Raw("SELECT sqlite_version()")
	if err := tx.Error; err != nil {
		return "", errors.New(err)
	}

	if err := tx.Scan(&version).Error; err != nil {
		return "", errors.New(err)
	}
	return version, nil
}

func Table(db *gorm.DB) (string, error) {
	var out string
	tables := make([]string, 0)
	tx := db.Raw("SELECT name FROM sqlite_master WHERE type='table' AND name != 'sqlite_sequence'")
	if err := tx.Error; err != nil {
		return "", errors.New(err)
	}

	if err := tx.Scan(&tables).Error; err != nil {
		return "", errors.New(err)
	}

	for _, table := range tables {
		out += fmt.Sprintln(table)
	}

	return out, nil
}

func sprint(keys []string, m []map[string]interface{}) (out string) {
	keyLengths := make(map[string]int)
	// max lengths
	for _, key := range keys {
		if len(fmt.Sprint(key)) > keyLengths[key] {
			keyLengths[key] = len(fmt.Sprint(key))
		}
	}
	for _, record := range m {
		for key, value := range record {
			if len(fmt.Sprint(value)) > keyLengths[key] {
				keyLengths[key] = len(fmt.Sprint(value))
			}
		}
	}

	// Print headers
	header := "|"
	for _, key := range keys {
		header += " " + key + strings.Repeat(" ", keyLengths[key]-len(key)) + " |"
	}

	out += fmt.Sprintln(header)

	// Print separator
	separator := "|"
	for _, key := range keys {
		separator += " " + strings.Repeat("-", keyLengths[key]) + " |"
	}
	out += fmt.Sprintln(separator)

	// Print data rows
	for _, record := range m {
		row := "|"
		for _, key := range keys {
			value := fmt.Sprint(record[key])
			row += " " + value + strings.Repeat(" ", keyLengths[key]-len(value)) + " |"
		}
		out += fmt.Sprintln(row)
	}
	return
}

func Select(db *gorm.DB, sql string) (string, error) {
	var results []map[string]interface{}
	rows, err := db.Raw(sql).Rows()
	if err != nil {
		return "", errors.New(err)
	}
	defer rows.Close()

	columns, _ := rows.Columns()
	values := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))
	for rows.Next() {
		for i := range columns {
			valuePtrs[i] = &values[i]
		}
		rows.Scan(valuePtrs...)
		entry := make(map[string]interface{})
		for i, col := range columns {
			var v interface{}
			val := values[i]
			b, ok := val.([]byte)
			if ok {
				v = string(b)
			} else {
				v = val
			}
			entry[col] = v
		}
		results = append(results, entry)
	}

	return sprint(columns, results), nil
}
