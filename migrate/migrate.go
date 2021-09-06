package migrate

import (
	"bufio"
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

// Logger function signature
type Logger func(s string, args ...interface{})

// StdLog is the log.Printf function from the standard library
var StdLog = log.Printf

// Migrate manages database migrations
type Migrate struct {
	db         *sql.DB
	versions   []int64
	logger     func(s string, args ...interface{})
	migrations map[int64]*Migration
}

// New creates a new Migrate with the given database and versions.
//
// If the provided logger function is not `nil` additional information will be logged during the
// migrations apply or discard.
func New(db *sql.DB, logger Logger, migrations map[int64]*Migration) (m *Migrate, err error) {
	if len(migrations) == 0 {
		return nil, fmt.Errorf("migrate: no migrations where provided")
	}
	m = &Migrate{}
	m.db = db
	m.migrations = make(map[int64]*Migration)
	m.migrations[0] = migration0
	m.versions = append(m.versions, 0)

	if logger == nil {
		logger = nopLogger
	}
	m.logger = logger

	for _, mig := range migrations {
		if mig.Version <= 0 {
			return nil, fmt.Errorf("migrate: migration version must be greater than 0")
		}
		m.migrations[mig.Version] = mig
		m.versions = append(m.versions, mig.Version)
	}

	sort.Slice(m.versions, func(i, j int) bool {
		return m.versions[i] < m.versions[j]
	})

	return m, nil
}

// NewWithFiles is like new but takes a fs.Fs as a source for migration files.
// Only files within the 1st level of the provided path matching the `(\d+)_(\w+)\.(apply|discard)\.sql`
// pattern will be added to the Migrate catalog.
func NewWithFiles(db *sql.DB, files fs.FS, logger Logger) (m *Migrate, err error) {
	migrations := make(map[int64]*Migration)
	if logger == nil {
		logger = nopLogger
	}

	// walk the provided fs.FS matching found 1st level files matching with the migrationRegexp
	// and adding them to the Migrate catalog
	fs.WalkDir(files, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// skip dirs
		if d.IsDir() {
			return nil
		}

		// skip nested files within the provided fs.Fs
		if path != d.Name() {
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

		source, err := fs.ReadFile(files, d.Name())
		if err != nil {
			return fmt.Errorf("migrate: error reading file: %s version: %w", d.Name(), err)
		}

		switch match[3] {
		case "apply":
			mig.Apply = string(source)
		case "discard":
			mig.Discard = string(source)
		}

		return nil
	})

	return New(db, logger, migrations)
}

// Version returns the current database migration version.
// If the database migrations are not initialized version is -1.
func (m *Migrate) Version(ctx context.Context) (version *Version, err error) {
	tx, err := m.db.BeginTx(ctx, options)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	return m.version(ctx, tx)
}

func (m *Migrate) version(ctx context.Context, tx *sql.Tx) (version *Version, err error) {
	row := tx.QueryRowContext(ctx, `
		SELECT version, date, name FROM migrations ORDER BY date DESC LIMIT 1
	`)

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
	return m.Apply(ctx, m.versions[len(m.versions)-1])
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
	if _, ok := m.migrations[version]; !ok && version != -1 {
		return fmt.Errorf("migrate: specified version: %d does not exist", version)
	}

	current, err := m.Version(ctx)
	if err != nil {
		return err
	}

	var versions []int64
	switch {
	case current.Version < version:
		versions = m.versions[current.Version+1 : version+1]

		for _, mig := range versions {
			if err := m.apply(ctx, m.migrations[mig], false); err != nil {
				return err
			}
		}

	case current.Version > version:
		versions = m.versions[version+1 : current.Version+1]

		for x := len(versions) - 1; x >= 0; x-- {
			if err := m.apply(ctx, m.migrations[versions[x]], true); err != nil {
				return err
			}
		}

	case current.Version == version:
		// return fmt.Errorf("migrate: already at version: %d", version)
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
		tx.Rollback()
		tx, err = m.db.BeginTx(ctx, options)
		if err != nil {
			return err
		}
	}

	var stmt string
	var raw string

	switch discard {
	case false:
		if mig.Version != current.Version+1 {
			return fmt.Errorf(
				"migrate: wrong sequence number, current: %d, proposed: %d, discard: %t",
				current.Version, mig.Version, discard)
		}
		raw = mig.Apply
	case true:
		if mig.Version != current.Version {
			return fmt.Errorf(
				"migrate: wrong sequence number, current: %d, proposed: %d, discard: %t",
				current.Version, mig.Version, discard)
		}
		raw = mig.Discard
	}

	if raw == "" {
		return nil
	}

	scanner := bufio.NewScanner(strings.NewReader(raw))
	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "--") {
			continue
		}

		if line[len(line)-1] == ';' {
			if stmt != "" {
				stmt += " "
			}
			stmt += string(line[:len(line)-1])

			m.logger("migrate: %s, discard: %t, statement: %s", mig.Name, discard, stmt)
			if _, err := tx.ExecContext(ctx, stmt); err != nil {
				return err
			}

			stmt = ""
			continue
		}

		if stmt != "" {
			stmt += " "
		}
		stmt += line
	}

	if stmt != "" {
		m.logger("migrate: %s, discard: %t, statement: %s", mig.Name, discard, stmt)
		if _, err := tx.ExecContext(ctx, stmt); err != nil {
			return err
		}
	}

	mig = m.migrations[mig.Version]
	if discard {
		mig = m.migrations[mig.Version-1]
	}
	if mig != nil {
		if err = m.set(ctx, tx, mig); err != nil {
			return err
		}
	}

	return tx.Commit()
}

// 0001_initial_schema.apply.sql
// 0001_initial_schema.discard.sql
var migrationRegexp = regexp.MustCompile(`(\d+)_(\w+)\.(apply|discard)\.sql`)
var options = &sql.TxOptions{Isolation: sql.LevelSerializable}

type Migration struct {
	Version int64
	Name    string
	Apply   string
	Discard string
}

type Version struct {
	Version int64
	Date    time.Time
	Name    string
}

var migration0 = &Migration{
	Version: 0,
	Name:    "create_migrations_table",
	Apply: `CREATE TABLE IF NOT EXISTS migrations (
		date timestamp NOT NULL,
		version bigint NOT NULL,
		name varchar(512) NOT NULL,
		PRIMARY KEY (date,version)
	)`,
	Discard: `DROP TABLE IF EXISTS migrations CASCADE`,
}

func nopLogger(_ string, _ ...interface{}) {}
