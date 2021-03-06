package statement

import (
	"fmt"
	"reflect"

	"github.com/brunotm/norm/internal/buffer"
	"github.com/brunotm/norm/internal/scan"
)

var (
	// ErrEmptyWithAlias will be returned when the a alias for a with clause is empty
	ErrEmptyWithAlias = fmt.Errorf("statement: empty with clause alias")

	// ErrInvalidArgNumber will be returned when there is a mismatch between placeholders and values for interpolation.
	ErrInvalidArgNumber = fmt.Errorf("statement: invalid number of arguments")
)

// Buffer represents the write buffer for building statements.
// Fits nicely with a strings.Builder or a bytes.Buffer.
type Buffer interface {
	WriteString(s string) (int, error)
	String() string
}

// Statement represents the statement builder interface.
type Statement interface {
	Build(Buffer) error
	String() (q string, err error)
}

// buildWhereIn builds a `WHERE` clause.
func buildWhere(buf Buffer, where []Statement) (err error) {
	for x := 0; x < len(where); x++ {
		if x == 0 {
			_, _ = buf.WriteString(" WHERE ")
		} else {
			_, _ = buf.WriteString(" AND ")
		}

		if err = where[x].Build(buf); err != nil {
			return err
		}
	}

	return nil
}

// InterfaceSlice converts any slice to a []interface{}
func InterfaceSlice(slice interface{}) []interface{} {
	s := reflect.ValueOf(slice)
	if s.Kind() != reflect.Slice {
		panic("InterfaceSlice() given a non-slice type")
	}

	// Keep the distinction between nil and empty slice input
	if s.IsNil() {
		return nil
	}

	ret := make([]interface{}, s.Len())

	for i := 0; i < s.Len(); i++ {
		ret[i] = s.Index(i).Interface()
	}

	return ret
}

// buildWhereIn builds a `WHERE IN (values)` clause.
func buildWhereIn(column string, values ...interface{}) (p *Part) {
	buf := buffer.New()
	defer buf.Release()

	p = &Part{}

	if len(values) == 1 && scan.IsSlice(values[0]) {
		values = InterfaceSlice(values[0])
	}

	_, _ = buf.WriteString(column)
	_, _ = buf.WriteString(" IN (")
	for x := 0; x < len(values); x++ {
		if x > 0 {
			_, _ = buf.WriteString(",")
		}
		_, _ = buf.WriteString("?")
		p.Values = append(p.Values, values[x])
	}
	_, _ = buf.WriteString(")")
	p.Query = buf.String()
	return p
}

// with represents a `WITH` clause.
type with struct {
	recursive bool
	alias     string
	stmt      Statement
}

// Build builds the statement into the given buffer.
func (s *with) Build(buf Buffer) (err error) {
	if s.alias == "" {
		return ErrEmptyWithAlias
	}

	var w string
	switch s.recursive {
	case false:
		w = "WITH "
	case true:
		w = "WITH RECURSIVE "
	}

	_, _ = buf.WriteString(w)
	_, _ = buf.WriteString(s.alias)
	_, _ = buf.WriteString(" AS (")
	if err = s.stmt.Build(buf); err != nil {
		return err
	}
	_, _ = buf.WriteString(")")
	return nil
}

// String builds the statement and returns the resulting query string.
func (s *with) String() (q string, err error) {
	buf := buffer.New()
	defer buf.Release()

	if err = s.Build(buf); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// union represents a `UNION` clause
type union struct {
	all  bool
	stmt Statement
}

// Build builds the statement into the given buffer.
func (s *union) Build(buf Buffer) (err error) {
	switch s.all {
	case false:
		_, _ = buf.WriteString("UNION ")
	case true:
		_, _ = buf.WriteString("UNION ALL ")
	}

	return s.stmt.Build(buf)
}

// String builds the statement and returns the resulting query string.
func (s *union) String() (q string, err error) {
	buf := buffer.New()
	defer buf.Release()

	if err = s.Build(buf); err != nil {
		return "", err
	}

	return buf.String(), nil
}
