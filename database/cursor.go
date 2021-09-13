package database

import (
	"database/sql"
	"fmt"
	"reflect"

	"github.com/brunotm/norm/internal/scan"
	"github.com/brunotm/norm/statement"
)

// Cursor is a cursor to a database result set.
type Cursor struct {
	rows      *sql.Rows
	vType     reflect.Type
	columns   []string
	extractor scan.PointersExtractor
}

// Scan copies the current row columns into the struct fields or map values pointed at by dst.
// If the type of dst changes during calls to scan it will return a error.
func (c *Cursor) Scan(dst interface{}) (err error) {
	v := reflect.ValueOf(dst)
	if v.Kind() != reflect.Ptr || v.IsNil() {
		return scan.ErrInvalidType
	}

	if c.extractor == nil {
		c.vType = v.Type()

		// cursor scan requires individual pointer to structs
		if v.Kind() == reflect.Slice && c.vType.Elem().Kind() != reflect.Uint8 {
			return scan.ErrInvalidType
		}

		if c.extractor, err = scan.FindExtractor(c.vType); err != nil {
			return err
		}
	}

	// check that we are not changing dst types during iteration
	if v.Type() != c.vType {
		return scan.ErrInvalidType
	}

	ptr := c.extractor(c.columns, v)
	err = c.rows.Scan(ptr...)
	if err != nil {
		return err
	}

	return nil
}

// Next prepares the next result row for reading with the Scan method.
// It returns true on success, or false if there is no next result row or an error happened while preparing it.
// Err should be consulted to distinguish between the two cases.
//
// Every call to Scan, even the first one, must be preceded by a call to Next.
func (c *Cursor) Next() (ok bool) {
	return c.rows.Next()
}

// Err returns the error, if any, that was encountered during iteration.
// Err may be called after an explicit or implicit Close.
func (c *Cursor) Err() (err error) {
	return c.rows.Err()
}

// Close closes the Cursor, preventing further enumeration.
// If Next is called and returns false and there are no further result sets,
// the Cursor is closed automatically and it will suffice to check the result of Err.
// Close is idempotent and does not affect the result of Err.
func (c *Cursor) Close() (err error) {
	return c.rows.Close()
}

// Cursor executes a query that returns a database cursor like sql.Rows.
// It its useful for working with large result sets or/and when memory utilization
// is a concern.
//
// The caller must call Cursor.Close() on the returned cursor in order to release
// the sql.Rows resources.
func (t *Tx) Cursor(stmt statement.Statement) (i *Cursor, err error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	query, err := stmt.String()
	if err != nil {
		return nil, err
	}

	r, err := t.tx.QueryContext(t.ctx, query)
	if err != nil {
		return nil, err
	}

	cursor := &Cursor{}
	cursor.rows = r
	if cursor.columns, err = r.Columns(); err != nil {
		return nil, fmt.Errorf("statement: %w", err)
	}

	return cursor, nil
}
