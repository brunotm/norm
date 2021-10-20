package migrate

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"log"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

var (
	// StdLog is the log.Printf function from the standard library
	StdLog = log.Printf

	// 0001_initial_schema.apply.sql
	// 0001_initial_schema.discard.sql
	migrationRegexp = regexp.MustCompile(`(\d+)_(\w+)\.(apply|discard)\.sql`)
	options         = &sql.TxOptions{Isolation: sql.LevelSerializable}

	versionQuery = "SELECT version, date, name FROM migrations ORDER BY date DESC LIMIT 1"

	migration0 = &Migration{
		Version: 0,
		Name:    "create_migrations_table",
		Apply: Statements{
			NoTx: false,
			Statements: []string{
				`CREATE TABLE IF NOT EXISTS migrations (date timestamp NOT NULL, version bigint NOT NULL, name varchar(512) NOT NULL, PRIMARY KEY (date,version))`},
		},
		Discard: Statements{
			NoTx:       false,
			Statements: []string{`DROP TABLE IF EXISTS migrations CASCADE`},
		},
	}
)

// Executor executes statements in a database
type Executor interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
}

// Logger function signature
type Logger func(s string, args ...interface{})

// nopLogger does notting
func nopLogger(_ string, _ ...interface{}) {}

// Migrate manages database migrations
type Migrate struct {
	db         *sql.DB
	logger     func(s string, args ...interface{})
	migrations []*Migration
}

// Migration represents a database migration apply and discard statements
type Migration struct {
	Version int64
	Name    string
	Apply   Statements
	Discard Statements
}

// Statements are set of SQL statements that either apply or discard a migration
type Statements struct {
	NoTx       bool
	Statements []string
}

// Version represents a migration version and its metadata
type Version struct {
	Version int64
	Date    time.Time
	Name    string
}

// New creates a new Migrate with the given database and versions.
//
// If the provided logger function is not `nil` additional information will be logged during the
// migrations apply or discard.
func New(db *sql.DB, logger Logger, migrations []*Migration) (m *Migrate, err error) {
	if len(migrations) == 0 {
		return nil, fmt.Errorf("migrate: no migrations where provided")
	}
	m = &Migrate{}
	m.db = db
	m.migrations = append(m.migrations, migration0)

	if logger == nil {
		logger = nopLogger
	}
	m.logger = logger

	for _, mig := range migrations {
		if mig.Version <= 0 {
			return nil, fmt.Errorf("migrate: migration version must be greater than 0")
		}

		m.migrations = append(m.migrations, mig)
	}

	sort.Slice(m.migrations, func(i, j int) bool {
		return m.migrations[i].Version < m.migrations[j].Version
	})

	// ensure migrations are sequential
	for x := 0; x < len(m.migrations); x++ {
		if m.migrations[x].Version != int64(x) {
			return nil, fmt.Errorf("migrate: migration versions must be sequential")
		}
	}

	return m, nil
}

// NewWithFiles is like new but takes a fs.Fs as a source for migration files.
// Only files within the 1st level of the provided path matching the `(\d+)_(\w+)\.(apply|discard)\.sql`
// pattern will be added to the Migrate catalog.
func NewWithFiles(db *sql.DB, logger Logger, files fs.FS) (m *Migrate, err error) {
	if logger == nil {
		logger = nopLogger
	}

	migrations := make(map[int64]*Migration)

	// walk the provided fs.FS matching found 1st level files matching with the migrationRegexp
	// and adding them to the Migrate catalog
	err = fs.WalkDir(files, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// skip dirs
		if d.IsDir() {
			return nil
		}

		match := migrationRegexp.FindStringSubmatch(d.Name())
		if len(match) != 4 {
			logger("migrate: could not match file in provided versions: %s, data: %#v", d.Name(), match)
			return nil
		}

		version, err := strconv.ParseInt(match[1], 10, 64)
		if err != nil {
			return fmt.Errorf("migrate: error parsing %#v version: %w", match, err)
		}

		if version <= 0 {
			return fmt.Errorf("migrate: migration version must be greater than 0")
		}

		mig, ok := migrations[version]
		if !ok {
			mig = &Migration{Version: version, Name: match[2]}
			migrations[version] = mig
		}
		logger("migrate: adding entry for: %s, file: %s", match[2], d.Name())

		source, err := fs.ReadFile(files, path)
		if err != nil {
			return fmt.Errorf("migrate: error reading file: %s version: %w", d.Name(), err)
		}

		switch match[3] {
		case "apply":
			mig.Apply, err = parseStatement(source)
		case "discard":
			mig.Discard, err = parseStatement(source)
		}

		return err
	})

	if err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}

	var arg []*Migration
	for _, m := range migrations {
		arg = append(arg, m)
	}

	return New(db, logger, arg)
}

// Version returns the current database migration version.
// If the database migrations are not initialized version is -1.
func (m *Migrate) Version(ctx context.Context) (version *Version, err error) {
	tx, err := m.db.BeginTx(ctx, options)
	if err != nil {
		return nil, err
	}

	version, err = m.version(ctx, tx)
	_ = tx.Rollback()

	return version, err
}

func (m *Migrate) version(ctx context.Context, tx *sql.Tx) (version *Version, err error) {
	row := tx.QueryRowContext(ctx, versionQuery)

	version = &Version{}
	err = row.Scan(&version.Version, &version.Date, &version.Name)

	switch {
	case err != nil && strings.Contains(strings.ToLower(err.Error()), "exist"):
		version.Version = -1
	case err != nil && err != sql.ErrNoRows:
		return nil, err
	}

	return version, nil
}

// Up apply all existing migrations to the database
func (m *Migrate) Up(ctx context.Context) (err error) {
	return m.Apply(ctx, m.migrations[len(m.migrations)-1].Version)
}

// Down discards all existing database migrations and migration history
func (m *Migrate) Down(ctx context.Context) (err error) {
	return m.Apply(ctx, -1)
}

func (m *Migrate) set(ctx context.Context, tx *sql.Tx, mig *Migration) (err error) {
	stmt := fmt.Sprintf(
		"INSERT INTO migrations(version, date, name) values(%d, NOW(), '%s')",
		mig.Version, mig.Name)

	m.logger(`migrate: update version, statement: %s`, stmt)
	_, err = tx.ExecContext(ctx, stmt)
	return err
}

// Apply either rolls forward or backwards the migrations to the specified version
func (m *Migrate) Apply(ctx context.Context, version int64) (err error) {
	if len(m.migrations) < int(version) && version != -1 {
		return fmt.Errorf("migrate: specified version: %d does not exist", version)
	}

	current, err := m.Version(ctx)
	if err != nil {
		return err
	}

	var migrations []*Migration
	switch {
	case current.Version < version:
		migrations = m.migrations[current.Version+1 : version+1]

		for mig := range migrations {
			if err := m.apply(ctx, m.migrations[mig], false); err != nil {
				return err
			}
		}

	case current.Version > version:
		migrations = m.migrations[version+1 : current.Version+1]

		for x := len(migrations) - 1; x >= 0; x-- {
			if err := m.apply(ctx, m.migrations[x], true); err != nil {
				return err
			}
		}

	case current.Version == version:
		return nil
	}

	return nil
}

func (m *Migrate) apply(ctx context.Context, mig *Migration, discard bool) (err error) {
	tx, err := m.db.BeginTx(ctx, options)
	if err != nil {
		return err
	}

	current, err := m.version(ctx, tx)
	if err != nil {
		return err
	}

	// restart tx if migrations are not initialized
	if current.Version == -1 {
		_ = tx.Rollback()
		tx, err = m.db.BeginTx(ctx, options)
		if err != nil {
			return err
		}
	}

	var statements Statements
	switch discard {
	case false:
		if mig.Version != current.Version+1 {
			return fmt.Errorf(
				"migrate: wrong sequence number, current: %d, proposed: %d, discard: %t",
				current.Version, mig.Version, discard)
		}
		statements = mig.Apply

	case true:
		if mig.Version != current.Version {
			return fmt.Errorf(
				"migrate: wrong sequence number, current: %d, proposed: %d, discard: %t",
				current.Version, mig.Version, discard)
		}
		statements = mig.Discard

	}

	for x := 0; x < len(statements.Statements); x++ {
		m.logger("migrate: %s, discard: %t, transaction: %t, statement: %s", mig.Name, discard, !statements.NoTx, statements.Statements[x])

		switch statements.NoTx {
		case false:
			if _, err := tx.ExecContext(ctx, statements.Statements[x]); err != nil {
				return err
			}

		case true:
			if _, err := m.db.ExecContext(ctx, statements.Statements[x]); err != nil {
				return err
			}
		}
	}

	// return early if we are discarding migration 0
	if mig.Version == 0 && discard {
		return tx.Commit()
	}

	// set the current version after applying the migration
	if discard {
		mig = m.migrations[mig.Version-1]
	}

	if err = m.set(ctx, tx, mig); err != nil {
		return err
	}

	return tx.Commit()
}
