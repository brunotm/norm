package statement

import (
	"fmt"
	"strings"
)

var (
	ErrEmptyWithAlias   = fmt.Errorf("statement: empty with clause alias")
	ErrInvalidArgNumber = fmt.Errorf("statement: invalid number of arguments")
)

// Buffer represents the write buffer for building statements.
// Fits nicely with a *strings.Builder.
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
			buf.WriteString(" WHERE ")
		} else {
			buf.WriteString(" AND ")
		}

		if err = where[x].Build(buf); err != nil {
			return err
		}
	}

	return nil
}

// buildWhereIn builds a `WHERE IN (values)` clause.
func buildWhereIn(column string, values ...interface{}) (p *part) {
	p = &part{}
	p.query = column + ` IN (`
	for x := 0; x < len(values); x++ {
		if x == 0 {
			p.query += "?"
		} else {
			p.query += ",?"
		}
		p.values = append(p.values, values[x])
	}
	p.query += `)`
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

	buf.WriteString(w + s.alias + " AS (")
	if err = s.stmt.Build(buf); err != nil {
		return err
	}
	buf.WriteString(")")
	return nil
}

// String builds the statement and returns the resulting query string.
func (s *with) String() (q string, err error) {
	var buf strings.Builder
	if err = s.Build(&buf); err != nil {
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
		buf.WriteString("UNION ")
	case true:
		buf.WriteString("UNION ALL ")
	}

	if err = s.stmt.Build(buf); err != nil {
		return err
	}

	return nil
}

// String builds the statement and returns the resulting query string.
func (s *union) String() (q string, err error) {
	var buf strings.Builder
	if err = s.Build(&buf); err != nil {
		return "", err
	}

	return buf.String(), nil
}
