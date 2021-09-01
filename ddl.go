package statement

import "strings"

// DDL represents a data definition statement.
type DDL struct {
	comment []Statement
	*part
}

// Comment adds a SQL comment to the generated query.
// Each call to comment creates a new `-- <comment>` line.
func (s *DDL) Comment(c string, values ...interface{}) *DDL {
	p := &part{}
	p.query = "-- " + c
	p.values = values
	s.comment = append(s.comment, p)
	return s
}

// Create creates a new `CREATE` DDL statement.
func Create(query string, values ...interface{}) *DDL {
	return &DDL{
		part: &part{
			query:  "CREATE " + query,
			values: values,
		},
	}
}

// Alter creates a new `ALTER` DDL statement.
func Alter(query string, values ...interface{}) *DDL {
	return &DDL{
		part: &part{
			query:  "ALTER " + query,
			values: values,
		},
	}
}

// Drop creates a new `DROP` DDL statement.
func Drop(query string, values ...interface{}) *DDL {
	return &DDL{
		part: &part{
			query:  "DROP " + query,
			values: values,
		},
	}
}

// Truncate creates a new `TRUNCATE` DDL statement.
func Truncate(query string, values ...interface{}) *DDL {
	return &DDL{
		part: &part{
			query:  "TRUNCATE " + query,
			values: values,
		},
	}
}

// Build builds the statement into the given buffer.
func (s *DDL) Build(buf Buffer) (err error) {
	for x := 0; x < len(s.comment); x++ {
		if err = s.comment[x].Build(buf); err != nil {
			return err
		}
		_, _ = buf.WriteString("\n")
	}
	return s.build(buf, true)
}

// String builds the statement and returns the resulting query string.
func (s *DDL) String() (q string, err error) {
	var buf strings.Builder
	if err = s.Build(&buf); err != nil {
		return "", err
	}

	return buf.String(), nil
}
