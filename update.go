package statement

import (
	"sort"
	"strings"
)

// UpdateStatement statement.
type UpdateStatement struct {
	table     string
	with      Statement
	values    map[string]interface{}
	where     []Statement
	returning []string
}

// Update creates a new update statement
func Update(table string) (s *UpdateStatement) {
	return &UpdateStatement{table: table, values: make(map[string]interface{})}
}

// Set adds a `SET column = value` clause, multiple calls to set append
// additional updates `SET column = value, column = value`
func (s *UpdateStatement) Set(column string, value interface{}) *UpdateStatement {
	s.values[column] = value
	return s
}

// SetMap specifies a map of column-value pairs to be updated.
func (s *UpdateStatement) SetMap(m map[string]interface{}) *UpdateStatement {
	for col, val := range m {
		s.values[col] = val
	}
	return s
}

// With adds a `WITH alias AS (stmt)` clause.
func (s *UpdateStatement) With(alias string, stmt Statement) *UpdateStatement {
	s.with = &with{alias: alias, stmt: stmt}
	return s
}

// Where adds a `WHERE` clause, multiple calls to Where are `ANDed` together.
func (s *UpdateStatement) Where(q string, values ...interface{}) *UpdateStatement {
	s.where = append(s.where, &part{query: q, values: values})
	return s
}

// WhereIn adds a `WHERE IN (values)` clause, multiple calls to WhereIn are `ANDed` together.
func (s *UpdateStatement) WhereIn(column string, values ...interface{}) *UpdateStatement {
	s.where = append(s.where, buildWhereIn(column, values...))
	return s
}

// Returning adds a `RETURNING columns` clause.
func (s *UpdateStatement) Returning(columns ...string) *UpdateStatement {
	s.returning = columns
	return s
}

// Build builds the statement into the given buffer.
func (s *UpdateStatement) Build(buf Buffer) (err error) {
	if s.with != nil {
		if err = s.with.Build(buf); err != nil {
			return err
		}
		buf.WriteString(" ")
	}

	buf.WriteString("UPDATE " + s.table)
	buf.WriteString(" SET")

	sorted := make([]string, 0, len(s.values))
	for k := range s.values {
		sorted = append(sorted, k)
	}
	sort.Strings(sorted)

	for x := 0; x < len(sorted); x++ {
		if x > 0 {
			buf.WriteString(", " + sorted[x] + " = ")
		} else {
			buf.WriteString(" " + sorted[x] + " = ")
		}
		writeValue(buf, s.values[sorted[x]], false)
	}

	if err = buildWhere(buf, s.where); err != nil {
		return err
	}

	if len(s.returning) > 0 {
		buf.WriteString(" RETURNING " + strings.Join(s.returning, ","))
	}

	return nil
}

// String builds the statement and returns the resulting query string.
func (s *UpdateStatement) String() (q string, err error) {
	var buf strings.Builder
	if err = s.Build(&buf); err != nil {
		return "", err
	}

	return buf.String(), nil
}
