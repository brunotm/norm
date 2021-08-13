package db

import (
	"context"
	"database/sql"
	"hash/maphash"
	"reflect"
	"sync"

	"github.com/brunotm/statement"
	"github.com/brunotm/statement/scan"
)

// DB is a wrapped *sql.DB
type DB struct {
	db       *sql.DB
	readOpt  *sql.TxOptions
	writeOpt *sql.TxOptions
}

// New creates a new database from an existing *sql.DB.
func New(db *sql.DB, read, write sql.IsolationLevel) (d *DB, err error) {
	d = &DB{}
	d.db = db

	d.readOpt = &sql.TxOptions{Isolation: read, ReadOnly: true}
	d.writeOpt = &sql.TxOptions{Isolation: write, ReadOnly: false}

	return d, nil
}

// Tx creates a database transaction with the provided options.
func (d *DB) Tx(ctx context.Context, opts *sql.TxOptions) (tx *Tx, err error) {
	t, err := d.db.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}

	return &Tx{
		tx:    t,
		ctx:   ctx,
		cache: map[uint64]reflect.Value{},
	}, nil

}

// Read creates a read-only transaction with the default DB isolation level.
func (d *DB) Read(ctx context.Context) (tx *Tx, err error) {
	return d.Tx(ctx, d.readOpt)
}

// Update creates a read-write transaction with the default DB isolation level.
func (d *DB) Update(ctx context.Context) (tx *Tx, err error) {
	return d.Tx(ctx, d.writeOpt)
}

// Tx represents a database transaction
type Tx struct {
	done  bool
	mu    sync.Mutex
	tx    *sql.Tx
	ctx   context.Context
	hash  maphash.Hash
	cache map[uint64]reflect.Value
}

// Exec executes a query that doesn't return rows.
func (t *Tx) Exec(stmt statement.Statement) (r sql.Result, err error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	query, err := stmt.String()
	if err != nil {
		return nil, err
	}

	return t.tx.ExecContext(t.ctx, query)
}

// Query executes a query that returns rows.
func (t *Tx) Query(dst interface{}, stmt statement.Statement) (err error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	query, err := stmt.String()
	if err != nil {
		return err
	}

	if _, err = t.hash.WriteString(query); err != nil {
		return err
	}

	key := t.hash.Sum64()
	t.hash.Reset()

	if r, ok := t.cache[key]; ok {
		reflect.ValueOf(dst).Elem().Set(r)
		return nil
	}

	r, err := t.tx.QueryContext(t.ctx, query)
	if err != nil {
		return err
	}

	if _, err = scan.Load(r, dst); err != nil {
		return err
	}

	if err == nil {
		t.cache[key] = reflect.ValueOf(dst).Elem()
	}

	return err
}

// Commit the transaction.
func (t *Tx) Commit() (err error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	err = t.tx.Commit()
	t.done = true

	return err
}

// Rollback aborts the transaction.
func (t *Tx) Rollback() (err error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.done {
		return nil
	}

	err = t.tx.Rollback()
	t.done = true
	return err
}
