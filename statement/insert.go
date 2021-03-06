package statement

import (
	"reflect"
	"sort"
	"strings"

	"github.com/brunotm/norm/internal/buffer"
	"github.com/brunotm/norm/internal/scan"
)

// InsertStatement statement.
type InsertStatement struct {
	table        string
	columns      []string
	values       []Statement
	comment      []Statement
	valuesSelect *SelectStatement
	with         Statement
	onConflict   Statement
	returning    []string
}

// Insert creates a new `INSERT` statement.
func Insert() (s *InsertStatement) {
	return &InsertStatement{}
}

// Comment adds a SQL comment to the generated query.
// Each call to comment creates a new `-- <comment>` line.
func (s *InsertStatement) Comment(c string, values ...interface{}) *InsertStatement {
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

// Into specifies the table on which to perform the insert
func (s *InsertStatement) Into(table string) (st *InsertStatement) {
	s.table = table
	return s
}

// Columns specifies the columns for the `INSERT` statement.
func (s *InsertStatement) Columns(columns ...string) (st *InsertStatement) {
	s.columns = columns
	return s
}

// Values specifies the values for the `VALUES` clause.
func (s *InsertStatement) Values(values ...interface{}) (st *InsertStatement) {
	p := &Part{}
	buf := buffer.New()
	defer buf.Release()

	_, _ = buf.WriteString("(")

	for x := 0; x < len(values); x++ {
		if x > 0 {
			_, _ = buf.WriteString(",")
		}
		_, _ = buf.WriteString("?")
		p.Values = append(p.Values, values[x])
	}
	_, _ = buf.WriteString(")")

	p.Query = buf.String()
	s.values = append(s.values, p)
	return s
}

// Record add the values from the given struct for insert.
// If no columns where specified before calling Record(), the columns will be defined by the struct fields.
func (s *InsertStatement) Record(structValue interface{}) (st *InsertStatement) {
	v := reflect.Indirect(reflect.ValueOf(structValue))

	if v.Kind() == reflect.Struct {
		var value []interface{}
		m := scan.StructMap(v.Type())

		// populate columns from available record fields
		// if no columns were specified up to this point
		if len(s.columns) == 0 {
			s.columns = make([]string, 0, len(m))
			for key := range m {
				s.columns = append(s.columns, key)
			}

			// ensure that the column ordering is deterministic
			sort.Strings(s.columns)
		}

		for _, key := range s.columns {
			if index, ok := m[key]; ok {
				value = append(value, v.FieldByIndex(index).Interface())
			} else {
				value = append(value, nil)
			}
		}
		s.Values(value...)
	}

	return s
}

// ValuesSelect specifies a Select statement from which values will be inserted.
func (s *InsertStatement) ValuesSelect(values *SelectStatement) (st *InsertStatement) {
	s.valuesSelect = values
	return s
}

// OnConflict adds a `ON CONFLICT` clause.
func (s *InsertStatement) OnConflict(q string, values ...interface{}) (st *InsertStatement) {
	buf := buffer.New()
	defer buf.Release()

	_, _ = buf.WriteString("ON CONFLICT ")
	_, _ = buf.WriteString(q)

	p := &Part{}
	p.Query = buf.String()
	p.Values = values

	s.onConflict = p
	return s
}

// With adds a `WITH alias AS (stmt)`
func (s *InsertStatement) With(alias string, stmt Statement) *InsertStatement {
	s.with = &with{alias: alias, stmt: stmt}
	return s
}

// Returning adds a `RETURNING columns` clause.
func (s *InsertStatement) Returning(columns ...string) *InsertStatement {
	s.returning = columns
	return s
}

// Build builds the statement into the given buffer.
func (s *InsertStatement) Build(buf Buffer) (err error) {
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

	_, _ = buf.WriteString("INSERT INTO ")
	_, _ = buf.WriteString(s.table)

	_, _ = buf.WriteString("(")
	_, _ = buf.WriteString(strings.Join(s.columns, ","))
	_, _ = buf.WriteString(")")

	if s.valuesSelect != nil {
		_, _ = buf.WriteString(" (")
		if err = s.valuesSelect.Build(buf); err != nil {
			return err
		}
		_, _ = buf.WriteString(")")
	} else {
		_, _ = buf.WriteString(" VALUES ")
		for x := 0; x < len(s.values); x++ {
			if err = s.values[0].Build(buf); err != nil {
				return err
			}
		}
	}

	if s.onConflict != nil {
		_, _ = buf.WriteString(" ")
		if err = s.onConflict.Build(buf); err != nil {
			return err
		}
	}

	if len(s.returning) > 0 {
		_, _ = buf.WriteString(" RETURNING ")
		_, _ = buf.WriteString(strings.Join(s.returning, ","))
	}

	return nil
}

// String builds the statement and returns the resulting query string.
func (s *InsertStatement) String() (q string, err error) {
	buf := buffer.New()
	defer buf.Release()

	if err = s.Build(buf); err != nil {
		return "", err
	}

	return buf.String(), nil
}
