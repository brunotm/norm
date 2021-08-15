package statement

import (
	"reflect"
	"sort"
	"strings"

	"github.com/brunotm/statement/scan"
)

// InsertStatement statement.
type InsertStatement struct {
	table        string
	columns      []string
	values       []Statement
	valuesSelect *SelectStatement
	with         Statement
	onConflict   Statement
	returning    []string
}

// Insert creates a new `INSERT` statement.
func Insert() (s *InsertStatement) {
	return &InsertStatement{}
}

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
	p := &part{}
	p.query = `(`
	for x := 0; x < len(values); x++ {
		if x == 0 {
			p.query += "?"
		} else {
			p.query += ",?"
		}
		p.values = append(p.values, values[x])
	}
	p.query += `)`

	s.values = append(s.values, p)
	return s
}

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

// OnConflictUpdate adds a `ON CONFLICT ON CONSTRAINT constraint DO UPDATE SET` clause.
func (s *InsertStatement) OnConflictUpdate(constraint string, actions map[string]interface{}) (st *InsertStatement) {
	p := &part{}
	p.query += `ON CONFLICT ON CONSTRAINT ` + constraint
	p.query += ` DO UPDATE SET`

	p.values = make([]interface{}, 0, len(actions))
	sorted := make([]string, 0, len(actions))

	for k := range actions {
		sorted = append(sorted, k)
	}
	sort.Strings(sorted)

	for x := 0; x < len(sorted); x++ {
		if x > 0 {
			p.query += `, ` + sorted[x] + ` = ?`
		} else {
			p.query += ` ` + sorted[x] + ` = ?`
		}
		p.values = append(p.values, actions[sorted[x]])
	}

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
	if s.with != nil {
		if err = s.with.Build(buf); err != nil {
			return err
		}
		buf.WriteString(" ")
	}

	buf.WriteString("INSERT INTO " + s.table)

	buf.WriteString("(")
	buf.WriteString(strings.Join(s.columns, ","))
	buf.WriteString(")")

	if s.valuesSelect != nil {
		buf.WriteString(" (")
		if err = s.valuesSelect.Build(buf); err != nil {
			return err
		}
		buf.WriteString(")")
	} else {
		buf.WriteString(" VALUES ")
		for x := 0; x < len(s.values); x++ {
			if err = s.values[0].Build(buf); err != nil {
				return err
			}
		}
	}

	if s.onConflict != nil {
		buf.WriteString(" ")
		if err = s.onConflict.Build(buf); err != nil {
			return err
		}
	}

	if len(s.returning) > 0 {
		buf.WriteString(" RETURNING " + strings.Join(s.returning, ","))
	}

	return nil
}

// String builds the statement and returns the resulting query string.
func (s *InsertStatement) String() (q string, err error) {
	var buf strings.Builder
	if err = s.Build(&buf); err != nil {
		return "", err
	}

	return buf.String(), nil
}
