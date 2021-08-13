package statement

import "strings"

type DDL struct {
	*part
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

// Create creates a new `ALTER` DDL statement.
func Alter(query string, values ...interface{}) *DDL {
	return &DDL{
		part: &part{
			query:  "ALTER " + query,
			values: values,
		},
	}
}

// Create creates a new `DROP` DDL statement.
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

func (s *DDL) Build(buf Buffer) (err error) {
	return s.build(buf, true)
}

func (s *DDL) String() (q string, err error) {
	var buf strings.Builder
	if err = s.Build(&buf); err != nil {
		return "", err
	}

	return buf.String(), nil
}
