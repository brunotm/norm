package statement

import (
	"fmt"
	"strings"

	"github.com/brunotm/norm/internal/buffer"
)

// Join types
type Join string

var (
	// InnerJoin type
	InnerJoin Join = "INNER JOIN"
	// LeftOuterJoin type
	LeftOuterJoin Join = "LEFT OUTER JOIN"
	// RightOuterJoin type
	RightOuterJoin Join = "RIGHT OUTER JOIN"
	// FullOuterJoin type
	FullOuterJoin Join = "FULL OUTER JOIN"
)

// SelectStatement statement.
type SelectStatement struct {
	limitCount     int64
	offsetCount    int64
	order          string
	isDistinct     bool
	isForUpdate    bool
	isSkipLocked   bool
	tableStatement bool
	with           Statement
	union          Statement
	table          Statement
	columns        []interface{}
	groupBy        []string
	orderBy        []string
	comment        []Statement
	join           []Statement
	where          []Statement
	having         []Statement
}

// Select creates a new `SELECT` statement.
func Select() *SelectStatement {
	return &SelectStatement{}
}

// Comment adds a SQL comment to the generated query.
// Each call to comment creates a new `-- <comment>` line.
func (s *SelectStatement) Comment(c string, values ...interface{}) *SelectStatement {
	buf := buffer.New()
	defer buf.Release()

	_, _ = buf.WriteString("-- ")
	_, _ = buf.WriteString(c)

	p := &Part{}
	p.Query = buf.String()
	p.Values = values
	s.comment = append(s.comment, p)
	return s
}

// Columns set the `SELECT` columns. Columns overwrites any previously set columns for this statement.
func (s *SelectStatement) Columns(columns ...interface{}) *SelectStatement {
	s.columns = columns
	return s
}

// Column append the given column to the `SELECT`. Column appends to the existing columns already specified.
// Used for more ellaborate column specification.
func (s *SelectStatement) Column(q string, values ...interface{}) *SelectStatement {
	s.columns = append(s.columns, &Part{Query: q, Values: values})
	return s
}

// From sets the table name or *Select statement for the `FROM` clause.
func (s *SelectStatement) From(table interface{}) *SelectStatement {
	switch table := table.(type) {
	case Statement:
		s.tableStatement = true
		s.table = table
	case string:
		s.table = &Part{Query: table}
	}

	return s
}

// Join adds a `JOIN ...` clause.
func (s *SelectStatement) Join(join Join, table, cond string, values ...interface{}) *SelectStatement {
	buf := buffer.New()
	defer buf.Release()

	_, _ = buf.WriteString(string(join))
	_, _ = buf.WriteString(" ")
	_, _ = buf.WriteString(table)
	_, _ = buf.WriteString(" ON ")
	_, _ = buf.WriteString(cond)

	p := &Part{}
	p.Values = values
	p.Query = buf.String()

	s.join = append(s.join, p)
	return s
}

// JoinInner adds a `INNER JOIN` clause.
func (s *SelectStatement) JoinInner(table, cond string, values ...interface{}) *SelectStatement {
	return s.Join(InnerJoin, table, cond, values...)
}

// JoinLeft adds a `LEFT OUTER JOIN` clause.
func (s *SelectStatement) JoinLeft(table, cond string, values ...interface{}) *SelectStatement {
	return s.Join(LeftOuterJoin, table, cond, values...)
}

// JoinRight adds a `RIGHT OUTER JOIN` clause.
func (s *SelectStatement) JoinRight(table, cond string, values ...interface{}) *SelectStatement {
	return s.Join(RightOuterJoin, table, cond, values...)
}

// JoinFull adds a `FULL OUTER JOIN` clause.
func (s *SelectStatement) JoinFull(table, cond string, values ...interface{}) *SelectStatement {
	return s.Join(FullOuterJoin, table, cond, values...)
}

// Where adds a `WHERE` clause, multiple calls to Where are `ANDed` together.
func (s *SelectStatement) Where(q string, values ...interface{}) *SelectStatement {
	s.where = append(s.where, &Part{Query: q, Values: values})
	return s
}

// Having adds a `HAVING` clause, multiple calls to Having are `ANDed` together.
func (s *SelectStatement) Having(q string, values ...interface{}) *SelectStatement {
	s.having = append(s.having, &Part{Query: q, Values: values})
	return s
}

// WhereIn adds a `WHERE IN (values)` clause, multiple calls to WhereIn are `ANDed` together.
func (s *SelectStatement) WhereIn(column string, values ...interface{}) *SelectStatement {
	s.where = append(s.where, buildWhereIn(column, values...))
	return s
}

// GroupBy adds a `GROUP BY columns` clause.
func (s *SelectStatement) GroupBy(columns ...string) *SelectStatement {
	s.groupBy = append(s.groupBy, columns...)
	return s
}

// OrderAsc adds a `ORDER BY columns ASC` clause.
func (s *SelectStatement) OrderAsc(columns ...string) *SelectStatement {
	s.orderBy = columns
	s.order = "ASC"
	return s
}

// OrderDesc adds a `ORDER BY columns DESC` clause.
func (s *SelectStatement) OrderDesc(columns ...string) *SelectStatement {
	s.orderBy = columns
	s.order = "DESC"
	return s
}

// Limit adds a `LIMIT n` clause.
func (s *SelectStatement) Limit(n int64) *SelectStatement {
	s.limitCount = n
	return s
}

// Offset adds a `OFFSET n` clause, only if LIMIT is also set.
func (s *SelectStatement) Offset(n int64) *SelectStatement {
	s.offsetCount = n
	return s
}

// Distinct adds a `DISTINCT` clause.
func (s *SelectStatement) Distinct() *SelectStatement {
	s.isDistinct = true
	return s
}

// ForUpdate a `FOR UPDATE` clause.
func (s *SelectStatement) ForUpdate() *SelectStatement {
	s.isForUpdate = true
	return s
}

// SkipLocked adds a `SKIP LOCKED` clause.
func (s *SelectStatement) SkipLocked() *SelectStatement {
	s.isSkipLocked = true
	return s
}

// With adds a `WITH alias AS (stmt)`
func (s *SelectStatement) With(alias string, stmt Statement) *SelectStatement {
	s.with = &with{recursive: false, alias: alias, stmt: stmt}
	return s
}

// WithRecursive adds a `WITH RECURSIVE alias AS (stmt)`
func (s *SelectStatement) WithRecursive(alias string, stmt Statement) *SelectStatement {
	s.with = &with{recursive: true, alias: alias, stmt: stmt}
	return s
}

// Union adds a `UNION` clause.
func (s *SelectStatement) Union(stmt Statement) *SelectStatement {
	s.union = &union{stmt: stmt}
	return s
}

// UnionAll adds a `UNION ALL` clause.
func (s *SelectStatement) UnionAll(stmt Statement) *SelectStatement {
	s.union = &union{all: true, stmt: stmt}
	return s
}

// Build builds the statement into the given buffer.
func (s *SelectStatement) Build(buf Buffer) (err error) {
	for x := 0; x < len(s.comment); x++ {
		if err = s.comment[x].Build(buf); err != nil {
			return err
		}
		_, _ = buf.WriteString("\n")
	}

	if s.with != nil {
		if err = s.with.Build(buf); err != nil {
			return err
		}
		_, _ = buf.WriteString(" ")
	}

	_, _ = buf.WriteString("SELECT ")

	if s.isDistinct {
		_, _ = buf.WriteString("DISTINCT ")
	}

	for x := 0; x < len(s.columns); x++ {
		if x > 0 {
			_, _ = buf.WriteString(`,`)
		}

		switch c := s.columns[x].(type) {
		case Statement:
			if err = c.Build(buf); err != nil {
				return err
			}

		case string:
			_, _ = buf.WriteString(c)
		}
	}

	if s.table != nil {
		_, _ = buf.WriteString(" FROM ")
		switch s.tableStatement {
		case true:
			_, _ = buf.WriteString(`( `)
			err = s.table.Build(buf)
			_, _ = buf.WriteString(` )`)
		case false:
			err = s.table.Build(buf)
		}

		if err != nil {
			return err
		}
	}

	for x := 0; x < len(s.join); x++ {
		_, _ = buf.WriteString(" ")
		err = s.join[x].Build(buf)
		if err != nil {
			return err
		}
	}

	if err = buildWhere(buf, s.where); err != nil {
		return err
	}

	if len(s.groupBy) > 0 {
		_, _ = buf.WriteString(" GROUP BY ")
		_, _ = buf.WriteString(strings.Join(s.groupBy, ","))
	}

	for x := 0; x < len(s.having); x++ {
		if x == 0 {
			_, _ = buf.WriteString(" HAVING ")
		} else {
			_, _ = buf.WriteString(" AND ")
		}

		if err = s.having[x].Build(buf); err != nil {
			return err
		}

	}

	if len(s.orderBy) > 0 {
		_, _ = buf.WriteString(" ORDER BY ")
		_, _ = buf.WriteString(strings.Join(s.orderBy, `,`))
		_, _ = buf.WriteString(" ")
		_, _ = buf.WriteString(s.order)
	}

	if s.limitCount > 0 {
		_, _ = buf.WriteString(fmt.Sprintf(" LIMIT %d OFFSET %d", s.limitCount, s.offsetCount))
	}

	if s.isForUpdate {
		_, _ = buf.WriteString(" FOR UPDATE")
	}

	if s.isSkipLocked {
		_, _ = buf.WriteString(" SKIP LOCKED")
	}

	if s.union != nil {
		_, _ = buf.WriteString(" ")
		if err = s.union.Build(buf); err != nil {
			return err
		}
	}

	return nil
}

// String builds the statement and returns the resulting query string.
func (s *SelectStatement) String() (q string, err error) {
	buf := buffer.New()
	defer buf.Release()

	if err = s.Build(buf); err != nil {
		return "", err
	}

	return buf.String(), nil
}
