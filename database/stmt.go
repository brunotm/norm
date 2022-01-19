package database

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/brunotm/norm/internal/scan"
)

type Stmt struct {
	tx   *Tx
	stmt *sql.Stmt
}

// Close closes the statement.
func (s *Stmt) Close() (err error) {
	start := time.Now()
	err = s.stmt.Close()

	s.tx.log("db.tx.stmt.close", s.tx.tid, err, time.Since(start), "")
	return err
}

// Exec executes a prepared statement with the given arguments and
// returns a Result summarizing the effect of the statement.
func (s *Stmt) Exec(args ...interface{}) (r sql.Result, err error) {
	start := time.Now()
	r, err = s.stmt.ExecContext(s.tx.ctx, args...)

	s.tx.log("db.tx.stmt.exec", s.tx.tid, err, time.Since(start), fmt.Sprintf("%+v", args))
	return r, err
}

// Query executes a prepared query statement with the given arguments
// and returns the query results as a *Rows.
func (s *Stmt) Query(dst interface{}, args ...interface{}) (err error) {
	start := time.Now()

	r, err := s.stmt.QueryContext(s.tx.ctx, args...)
	if err != nil {
		return err
	}
	defer r.Close()

	_, err = scan.Load(r, dst)
	s.tx.log("db.tx.stmt.query", s.tx.tid, err, time.Since(start), fmt.Sprintf("%+v", args))
	return err

}
