package database

import (
	"context"
	"database/sql"
	"fmt"
	"hash/maphash"
	"reflect"
	"sync"
	"time"

	"github.com/brunotm/norm/internal/scan"
	"github.com/brunotm/norm/statement"
)

// Tx represents a database transaction
type Tx struct {
	mu    sync.Mutex
	tid   string
	log   Logger
	done  bool
	tx    *sql.Tx
	ctx   context.Context
	hash  maphash.Hash
	cache map[uint64]reflect.Value
}

// Prepare creates a prepared statement for use within a transaction.
func (t *Tx) Prepare(query string) (stmt *Stmt, err error) {
	start := time.Now()

	s, err := t.tx.PrepareContext(t.ctx, query)
	t.log("db.tx.prepare", t.tid, err, time.Since(start), query)
	if err != nil {
		return nil, err
	}

	return &Stmt{tx: t, stmt: s}, err
}

// Exec executes a query that doesn't return rows.
func (t *Tx) Exec(stmt statement.Statement) (r sql.Result, err error) {
	start := time.Now()
	t.mu.Lock()
	defer t.mu.Unlock()

	query, err := stmt.String()
	if err != nil {
		return nil, err
	}

	r, err = t.tx.ExecContext(t.ctx, query)

	t.log("db.tx.exec", t.tid, err, time.Since(start), query)
	return r, err
}

// ExecSQL is like Exec but accepts a raw SQL statement and values for interpolation
func (t *Tx) ExecSQL(query string, values ...interface{}) (r sql.Result, err error) {
	stmt := &statement.Part{Query: query, Values: values}
	return t.Exec(stmt)
}

// Query executes a query that returns rows.
func (t *Tx) Query(dst interface{}, stmt statement.Statement) (err error) {
	return t.query(dst, stmt, false)
}

// QuerySQL is like Query but accepts a raw SQL statement and values for interpolation
func (t *Tx) QuerySQL(dst interface{}, query string, values ...interface{}) (err error) {
	stmt := &statement.Part{Query: query, Values: values}
	return t.query(dst, stmt, false)
}

// QueryCache is like Query, but will add query results to or return already cached
// results from the transaction query cache.
func (t *Tx) QueryCache(dst interface{}, stmt statement.Statement) (err error) {
	return t.query(dst, stmt, true)
}

// QueryCacheSQL is like QueryCache but accepts a raw SQL statement and values for interpolation
func (t *Tx) QueryCacheSQL(dst interface{}, query string, values ...interface{}) (err error) {
	stmt := &statement.Part{Query: query, Values: values}
	return t.query(dst, stmt, true)
}

func (t *Tx) query(dst interface{}, stmt statement.Statement, cache bool) (err error) {
	start := time.Now()

	query, err := stmt.String()
	if err != nil {
		return err
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	var key uint64
	if cache {
		if _, err = t.hash.WriteString(query); err != nil {
			return err
		}

		key = t.hash.Sum64()
		t.hash.Reset()

		if r, ok := t.cache[key]; ok {
			dstValue := reflect.ValueOf(dst)

			if dstValue.Kind() != reflect.Ptr {
				err := fmt.Errorf("database: dst must be a pointer type")
				t.log("db.tx.query.cache.get", t.tid, err, time.Since(start), query)
				return err
			}

			if dstValue.Elem().Type() != r.Type() {
				err := fmt.Errorf("database: invalid cached dst type: %s, expected: %s",
					dstValue.Type().String(), r.Type().String())
				t.log("db.tx.query.cache.get", t.tid, err, time.Since(start), query)
				return err
			}

			dstValue.Elem().Set(r)
			t.log("db.tx.query.cache.get", t.tid, nil, time.Since(start), query)
			return nil
		}
	}

	r, err := t.tx.QueryContext(t.ctx, query)
	if err != nil {
		t.log("db.tx.query", t.tid, err, time.Since(start), query)
		return err
	}
	defer r.Close()

	if _, err = scan.Load(r, dst); err != nil {
		t.log("db.tx.query", t.tid, err, time.Since(start), query)
		return err
	}

	if cache {
		t.cache[key] = reflect.ValueOf(dst).Elem()
		t.log("db.tx.query.cache.add", t.tid, nil, time.Since(start), query)
	} else {
		t.log("db.tx.query", t.tid, err, time.Since(start), query)
	}

	return nil
}

// Commit the transaction.
func (t *Tx) Commit() (err error) {
	start := time.Now()
	t.mu.Lock()
	defer t.mu.Unlock()

	err = t.tx.Commit()
	t.done = true

	t.log("db.tx.commit", t.tid, err, time.Since(start), "")
	return err
}

// Rollback aborts the transaction.
func (t *Tx) Rollback() (err error) {
	start := time.Now()
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.done {
		return nil
	}

	err = t.tx.Rollback()
	t.done = true

	t.log("db.tx.rollback", t.tid, err, time.Since(start), "")
	return err
}
