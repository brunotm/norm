package statement

import (
	"fmt"
	"strings"
)

type Join string

var (
	InnerJoin      Join = "INNER JOIN"
	LeftOuterJoin  Join = "LEFT OUTER JOIN"
	RightOuterJoin Join = "RIGHT OUTER JOIN"
	FullOuterJoin  Join = "FULL OUTER JOIN"
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
	columns        []string
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
	p := &part{}
	p.query = "-- " + c
	p.values = values
	s.comment = append(s.comment, p)
	return s
}

// Columns set the `SELECT` columns.
func (s *SelectStatement) Columns(columns ...string) *SelectStatement {
	s.columns = columns
	return s
}

// From sets the table name or *Select statement for the `FROM` clause.
func (s *SelectStatement) From(table interface{}) *SelectStatement {
	switch table := table.(type) {
	case Statement:
		s.tableStatement = true
		s.table = table
	case string:
		s.table = &part{query: table}
	}

	return s
}

// Join adds a `JOIN ...` clause.
func (s *SelectStatement) Join(join Join, table, cond string, values ...interface{}) *SelectStatement {
	p := &part{}
	p.values = values
	p.query = string(join) + " " + table + " ON " + cond

	s.join = append(s.join, p)
	return s
}

// Join adds a `INNER JOIN` clause.
func (s *SelectStatement) JoinInner(table, cond string, values ...interface{}) *SelectStatement {
	return s.Join(InnerJoin, table, cond, values...)
}

// Join adds a `LEFT OUTER JOIN` clause.
func (s *SelectStatement) JoinLeft(table, cond string, values ...interface{}) *SelectStatement {
	return s.Join(LeftOuterJoin, table, cond, values...)
}

// Join adds a `RIGHT OUTER JOIN` clause.
func (s *SelectStatement) JoinRight(table, cond string, values ...interface{}) *SelectStatement {
	return s.Join(RightOuterJoin, table, cond, values...)
}

// Join adds a `FULL OUTER JOIN` clause.
func (s *SelectStatement) JoinFull(table, cond string, values ...interface{}) *SelectStatement {
	return s.Join(FullOuterJoin, table, cond, values...)
}

// Where adds a `WHERE` clause, multiple calls to Where are `ANDed` together.
func (s *SelectStatement) Where(q string, values ...interface{}) *SelectStatement {
	s.where = append(s.where, &part{query: q, values: values})
	return s
}

// Having adds a `HAVING` clause, multiple calls to Having are `ANDed` together.
func (s *SelectStatement) Having(q string, values ...interface{}) *SelectStatement {
	s.having = append(s.having, &part{query: q, values: values})
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
		buf.WriteString("\n")
	}

	if s.with != nil {
		if err = s.with.Build(buf); err != nil {
			return err
		}
		buf.WriteString(" ")
	}

	buf.WriteString("SELECT ")

	if s.isDistinct {
		buf.WriteString("DISTINCT ")
	}

	buf.WriteString(strings.Join(s.columns, ","))

	if s.table != nil {
		buf.WriteString(" FROM ")
		switch s.tableStatement {
		case true:
			buf.WriteString(`( `)
			err = s.table.Build(buf)
			buf.WriteString(` )`)
		case false:
			err = s.table.Build(buf)
		}

		if err != nil {
			return err
		}
	}

	for x := 0; x < len(s.join); x++ {
		buf.WriteString(" ")
		err = s.join[x].Build(buf)
		if err != nil {
			return err
		}
	}

	if err = buildWhere(buf, s.where); err != nil {
		return err
	}

	if len(s.groupBy) > 0 {
		buf.WriteString(" GROUP BY ")
		buf.WriteString(strings.Join(s.groupBy, "', '"))
	}

	for x := 0; x < len(s.having); x++ {
		if x == 0 {
			buf.WriteString(" HAVING ")
		} else {
			buf.WriteString(" AND ")
		}

		if err = s.having[x].Build(buf); err != nil {
			return err
		}

	}

	if len(s.orderBy) > 0 {
		buf.WriteString(" ORDER BY ")
		buf.WriteString(strings.Join(s.orderBy, `,`))
		buf.WriteString(" " + s.order)
	}

	if s.limitCount > 0 {
		buf.WriteString(fmt.Sprintf(" LIMIT %d OFFSET %d", s.limitCount, s.offsetCount))
	}

	if s.isForUpdate {
		buf.WriteString(" FOR UPDATE")
	}

	if s.isSkipLocked {
		buf.WriteString(" SKIP LOCKED")
	}

	if s.union != nil {
		buf.WriteString(" ")
		if err = s.union.Build(buf); err != nil {
			return err
		}
	}

	return nil
}

// String builds the statement and returns the resulting query string.
func (s *SelectStatement) String() (q string, err error) {
	var buf strings.Builder
	if err = s.Build(&buf); err != nil {
		return "", err
	}

	return buf.String(), nil
}
