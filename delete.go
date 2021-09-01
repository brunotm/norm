package statement

import (
	"strings"
)

// DeleteStatement statement.
type DeleteStatement struct {
	table     string
	with      Statement
	comment   []Statement
	where     []Statement
	returning []string
}

// Delete creates a new `DELETE` statement.
func Delete() (s *DeleteStatement) {
	return &DeleteStatement{}
}

// Comment adds a SQL comment to the generated query.
// Each call to comment creates a new `-- <comment>` line.
func (s *DeleteStatement) Comment(c string, values ...interface{}) *DeleteStatement {
	p := &part{}
	p.query = "-- " + c
	p.values = values
	s.comment = append(s.comment, p)
	return s
}

// From sets the table name or for the `FROM` clause.
func (s *DeleteStatement) From(table string) *DeleteStatement {
	s.table = table
	return s
}

// With adds a `WITH alias AS (stmt)`
func (s *DeleteStatement) With(alias string, stmt Statement) *DeleteStatement {
	s.with = &with{alias: alias, stmt: stmt}
	return s
}

// Where adds a `WHERE` clause, multiple calls to Where are `ANDed` together.
func (s *DeleteStatement) Where(q string, values ...interface{}) *DeleteStatement {
	s.where = append(s.where, &part{query: q, values: values})
	return s
}

// WhereIn adds a `WHERE IN (values)` clause, multiple calls to WhereIn are `ANDed` together.
func (s *DeleteStatement) WhereIn(column string, values ...interface{}) *DeleteStatement {
	s.where = append(s.where, buildWhereIn(column, values...))
	return s
}

// Returning adds a `RETURNING columns` clause.
func (s *DeleteStatement) Returning(columns ...string) *DeleteStatement {
	s.returning = columns
	return s
}

// Build builds the statement into the given buffer.
func (s *DeleteStatement) Build(buf Buffer) (err error) {
	for x := 0; x < len(s.comment); x++ {
		if err = s.comment[x].Build(buf); err != nil {
			return err
		}
		buf.WriteString("\n")
	}

	if s.with != nil {
		if err = s.with.Build(buf); err != nil {
			return err
		}
		buf.WriteString(" ")
	}

	buf.WriteString("DELETE FROM " + s.table)
	if err = buildWhere(buf, s.where); err != nil {
		return err
	}

	if len(s.returning) > 0 {
		buf.WriteString(" RETURNING " + strings.Join(s.returning, ","))
	}

	return nil
}

// String builds the statement and returns the resulting query string.
func (s *DeleteStatement) String() (q string, err error) {
	var buf strings.Builder
	if err = s.Build(&buf); err != nil {
		return "", err
	}

	return buf.String(), nil
}
