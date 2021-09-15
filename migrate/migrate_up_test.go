package migrate

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestMigrationUp(t *testing.T) {
	mdb, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("error opening mock database: %s", err)
	}
	defer mdb.Close()

	// initial version check, version check returns relation does not exist error
	mock.ExpectBegin()
	mock.ExpectQuery(versionQuery).WillReturnError(fmt.Errorf("relation does not exist"))
	mock.ExpectRollback()

	// initial version check for migration0, relation does not exist
	mock.ExpectBegin()
	mock.ExpectQuery(versionQuery).WillReturnError(fmt.Errorf("relation does not exist"))
	mock.ExpectRollback()
	mock.ExpectBegin()
	mock.ExpectExec(migration0.Apply).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`INSERT INTO migrations(version, date, name) values(0, NOW(), 'create_migrations_table')`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	// initial version check for migration1, version check returns 0
	mock.ExpectBegin()
	mock.ExpectQuery(versionQuery).WillReturnRows(
		sqlmock.NewRows([]string{"date", "version", "name"}).
			AddRow(migration0.Version, time.Now(), migration0.Name),
	)
	mock.ExpectExec(migration1.Apply).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`INSERT INTO migrations(version, date, name) values(1, NOW(), 'users_table')`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	// initial version check for migration2, version check returns 1
	mock.ExpectBegin()
	mock.ExpectQuery(versionQuery).WillReturnRows(
		sqlmock.NewRows([]string{"date", "version", "name"}).
			AddRow(migration1.Version, time.Now(), migration1.Name),
	)
	mock.ExpectExec(migration2.Apply).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`INSERT INTO migrations(version, date, name) values(2, NOW(), 'users_email_index')`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	// initial version check for migration3, version check returns 2
	mock.ExpectBegin()
	mock.ExpectQuery(versionQuery).WillReturnRows(
		sqlmock.NewRows([]string{"date", "version", "name"}).
			AddRow(migration2.Version, time.Now(), migration2.Name),
	)
	mock.ExpectExec(migration3.Apply).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`INSERT INTO migrations(version, date, name) values(3, NOW(), 'roles_table')`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	// initial version check for migration4, version check returns 3
	mock.ExpectBegin()
	mock.ExpectQuery(versionQuery).WillReturnRows(
		sqlmock.NewRows([]string{"date", "version", "name"}).
			AddRow(migration3.Version, time.Now(), migration3.Name),
	)
	mock.ExpectExec(migration4.Apply).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`INSERT INTO migrations(version, date, name) values(4, NOW(), 'user_roles_fk')`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	m, err := New(mdb, StdLog, migrations)
	if err != nil {
		t.Fatalf("failed to create migrate: %s", err)
	}

	if err := m.Up(context.Background()); err != nil {
		t.Fatalf("migration run failed: %s", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("mock expectations failed: %s", err)
	}
}

var (
	migrations = []*Migration{migration4, migration3, migration2, migration1}

	migration1 = &Migration{
		Version: 1,
		Name:    "users_table",
		Apply:   "CREATE TABLE IF NOT EXISTS users(id text, name text, email text, role text, PRIMARY KEY (id))",
		Discard: "DROP TABLE IF EXISTS users CASCADE",
	}
	migration2 = &Migration{
		Version: 2,
		Name:    "users_email_index",
		Apply:   "CREATE INDEX IF NOT EXISTS ix_users_email ON users (email)",
		Discard: "DROP INDEX IF EXISTS ix_users_email CASCADE",
	}
	migration3 = &Migration{
		Version: 3,
		Name:    "roles_table",
		Apply:   "CREATE TABLE IF NOT EXISTS roles(id text, name text, properties jsonb NOT NULL DEFAULT '{}'::jsonb, PRIMARY KEY (id))",
		Discard: "DROP TABLE IF EXISTS roles CASCADE",
	}
	migration4 = &Migration{
		Version: 4,
		Name:    "user_roles_fk",
		Apply:   "ALTER TABLE users ADD CONSTRAINT roles_fk FOREIGN KEY (role) REFERENCES roles (id)",
		Discard: "ALTER TABLE users DROP CONSTRAINT roles_fk CASCADE",
	}
)
