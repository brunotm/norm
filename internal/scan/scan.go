package scan

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"unicode"
)

var (
	ErrInvalidType = fmt.Errorf("statement: invalid type for scan")
	typeValuer     = reflect.TypeOf((*driver.Valuer)(nil)).Elem()
	structMapCache = sync.Map{} // reflect.Type / map[string][]int
)

// IsSlice return true if the given interface{} holds a slice type
func IsSlice(v interface{}) bool {
	kind := reflect.TypeOf(v).Kind()
	if reflect.TypeOf(v).Kind() == reflect.Ptr {
		return reflect.Indirect(reflect.ValueOf(v)).Kind() == reflect.Slice
	}

	return kind == reflect.Slice
}

// Scan code adapted from https://github.com/mailru/dbr/blob/master/load.go

// Load loads any value from sql.Rows
func Load(rows *sql.Rows, value interface{}) (int, error) {
	defer rows.Close()
	var count int

	column, err := rows.Columns()
	if err != nil {
		return 0, err
	}

	v := reflect.ValueOf(value)
	if v.Kind() != reflect.Ptr || v.IsNil() {
		return 0, ErrInvalidType
	}

	v = v.Elem()
	isSlice := v.Kind() == reflect.Slice && v.Type().Elem().Kind() != reflect.Uint8

	var elemType reflect.Type
	if isSlice {
		elemType = v.Type().Elem()
	} else {
		elemType = v.Type()
	}

	extractor, err := FindExtractor(elemType)
	if err != nil {
		return count, err
	}

	for rows.Next() {
		var elem reflect.Value

		if isSlice {
			elem = reflect.New(elemType).Elem()
		} else {
			elem = v
		}

		ptr := extractor(column, elem)

		err = rows.Scan(ptr...)
		if err != nil {
			return count, err
		}
		count++

		if isSlice {
			v.Set(reflect.Append(v, elem))
		} else {
			break
		}
	}

	return count, rows.Err()
}

type dummyScanner struct{}

func (dummyScanner) Scan(interface{}) error {
	return nil
}

type keyValueMap map[string]interface{}

type kvScanner struct {
	column string
	m      keyValueMap
}

func (kv *kvScanner) Scan(v interface{}) error {
	if b, ok := v.([]byte); ok {
		tmp := make([]byte, len(b))
		copy(tmp, b)
		kv.m[kv.column] = tmp
	} else {
		// int64, float64, bool, string, time.Time, nil
		kv.m[kv.column] = v
	}
	return nil
}

// PointersExtractor function type
type PointersExtractor func(columns []string, value reflect.Value) []interface{}

var (
	dummyDest       sql.Scanner = dummyScanner{}
	typeScanner                 = reflect.TypeOf((*sql.Scanner)(nil)).Elem()
	typeKeyValueMap             = reflect.TypeOf(keyValueMap(nil))
)

func getStructFieldsExtractor(t reflect.Type) PointersExtractor {
	mapping := StructMap(t)
	return func(columns []string, value reflect.Value) []interface{} {
		var ptr []interface{}
		for _, key := range columns {
			if index, ok := mapping[key]; ok {
				ptr = append(ptr, value.FieldByIndex(index).Addr().Interface())
			} else {
				ptr = append(ptr, dummyDest)
			}
		}
		return ptr
	}
}

func getIndirectExtractor(extractor PointersExtractor) PointersExtractor {
	return func(columns []string, value reflect.Value) []interface{} {
		if value.IsNil() {
			value.Set(reflect.New(value.Type().Elem()))
		}
		return extractor(columns, value.Elem())
	}
}

func mapExtractor(columns []string, value reflect.Value) []interface{} {
	if value.IsNil() {
		value.Set(reflect.MakeMap(value.Type()))
	}
	m := value.Convert(typeKeyValueMap).Interface().(keyValueMap)
	var ptr = make([]interface{}, 0, len(columns))
	for _, c := range columns {
		ptr = append(ptr, &kvScanner{column: c, m: m})
	}
	return ptr
}

func dummyExtractor(columns []string, value reflect.Value) []interface{} {
	return []interface{}{value.Addr().Interface()}
}

// FindExtractor returns a PointersExtractor for the given type
func FindExtractor(t reflect.Type) (PointersExtractor, error) {
	if reflect.PtrTo(t).Implements(typeScanner) {
		return dummyExtractor, nil
	}

	switch t.Kind() {
	case reflect.Map:
		if !t.ConvertibleTo(typeKeyValueMap) {
			return nil, fmt.Errorf("statement: expected %v, got %v", typeKeyValueMap, t)
		}
		return mapExtractor, nil
	case reflect.Ptr:
		inner, err := FindExtractor(t.Elem())
		if err != nil {
			return nil, err
		}
		return getIndirectExtractor(inner), nil
	case reflect.Struct:
		return getStructFieldsExtractor(t), nil
	}

	return dummyExtractor, nil
}

// StructMap builds index to fast lookup fields in struct
func StructMap(t reflect.Type) map[string][]int {
	if m, _ := structMapCache.Load(t); m != nil {
		return m.(map[string][]int)
	}

	m := make(map[string][]int)
	structTraverse(m, t, nil)
	return m
}

func structTraverse(m map[string][]int, t reflect.Type, head []int) {
	if t.Implements(typeValuer) {
		return
	}
	switch t.Kind() {
	case reflect.Ptr:
		structTraverse(m, t.Elem(), head)
	case reflect.Struct:
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			if field.PkgPath != "" && !field.Anonymous {
				continue // not exported
			}
			tag := field.Tag.Get("db")
			if tag == "-" {
				continue // ignore
			}
			if tag == "" {
				// no tag, but we can record the field name
				tag = camelCaseToSnakeCase(field.Name)
			}
			if _, ok := m[tag]; !ok {
				m[tag] = append(head, i)
			}
			structTraverse(m, field.Type, append(head, i))
		}
	}
}

func camelCaseToSnakeCase(name string) string {
	var buf strings.Builder

	runes := []rune(name)

	for i := 0; i < len(runes); i++ {
		buf.WriteRune(unicode.ToLower(runes[i]))
		if i != len(runes)-1 && unicode.IsUpper(runes[i+1]) &&
			(unicode.IsLower(runes[i]) || unicode.IsDigit(runes[i]) ||
				(i != len(runes)-2 && unicode.IsLower(runes[i+2]))) {
			buf.WriteRune('_')
		}
	}

	return buf.String()
}
