package database

import (
	"context"
	"database/sql"
	"log"
	"reflect"
	"strconv"
	"time"
)

// Logger type for database operations
type Logger func(id, message string, err error, d time.Duration, query string)

// DefaultLogger for database operations
func DefaultLogger(id, message string, err error, d time.Duration, query string) {
	log.Printf("id: %s, message: %s, error: %s, duration_millis: %d, query: %s",
		id, message, err, d.Milliseconds(), query)
}

func nopLogger(message, id string, err error, d time.Duration, query string) {}

// DB is safe sql.DB wrapper which enforces transactional access to the database,
// transaction query caching and operation logging and plays nicely with `noorm/statement`.
type DB struct {
	db       *sql.DB
	log      Logger
	readOpt  *sql.TxOptions
	writeOpt *sql.TxOptions
}

// New creates a new database from an existing *sql.DB
// with the given sql.IsolationLevel and logger.
func New(db *sql.DB, level sql.IsolationLevel, logger Logger) (d *DB, err error) {
	d = &DB{}
	d.db = db
	d.log = nopLogger

	if logger != nil {
		d.log = logger
	}

	d.readOpt = &sql.TxOptions{Isolation: level, ReadOnly: true}
	d.writeOpt = &sql.TxOptions{Isolation: level, ReadOnly: false}

	return d, nil
}

// Tx creates a database transaction with the provided options.
// The tid argument is the transaction identifier that will be used to log operations
// done within the transaction.
func (d *DB) Tx(ctx context.Context, tid string, opts *sql.TxOptions) (tx *Tx, err error) {
	t, err := d.db.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}

	if tid == "" {
		tid = strconv.FormatInt(time.Now().UnixNano(), 32)
	}

	return &Tx{
		tid:   tid,
		log:   d.log,
		tx:    t,
		ctx:   ctx,
		cache: map[uint64]reflect.Value{},
	}, nil

}

// Read creates a read-only transaction with the default DB isolation level.
// The tid argument is the transaction identifier that will be used to log operations
// done within the transaction.
func (d *DB) Read(ctx context.Context, tid string) (tx *Tx, err error) {
	return d.Tx(ctx, tid, d.readOpt)
}

// Update creates a read-write transaction with the default DB isolation level.
// The tid argument is the transaction identifier that will be used to log operations
// done within the transaction.
func (d *DB) Update(ctx context.Context, tid string) (tx *Tx, err error) {
	return d.Tx(ctx, tid, d.writeOpt)
}
