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
type Logger func(message, tid string, err error, d time.Duration, query string)

// DefaultLogger for database operations
func DefaultLogger(message, tid string, err error, d time.Duration, query string) {
	log.Printf("message: %s, tid: %s, error: %s, duration_millis: %d, query: %s",
		message, tid, err, d.Milliseconds(), query)
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
	if tid == "" {
		tid = strconv.FormatInt(time.Now().UnixNano(), 32)
	}

	start := time.Now()
	t, err := d.db.BeginTx(ctx, opts)
	d.log("db.begin", tid, err, time.Since(start), "")

	if err != nil {
		return nil, err
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

// PingContext verifies a connection to the database is still alive,
// establishing a connection if necessary.
func (d *DB) Ping(ctx context.Context) (err error) {
	return d.db.PingContext(ctx)
}

// Close closes the database and prevents new queries from starting.
// Close then waits for all queries that have started processing on the server to finish.
func (d *DB) Close() (err error) {
	return d.db.Close()
}
