package migrate

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestMigrationDown(t *testing.T) {
	mdb, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("error opening mock database: %s", err)
	}
	defer mdb.Close()

	// initial version check, version check returns migration version 4
	mock.ExpectBegin()
	mock.ExpectQuery(versionQuery).WillReturnRows(
		sqlmock.NewRows([]string{"date", "version", "name"}).
			AddRow(migration4.Version, time.Now(), migration4.Name),
	)
	mock.ExpectRollback()

	mock.ExpectBegin()
	mock.ExpectQuery(versionQuery).WillReturnRows(
		sqlmock.NewRows([]string{"date", "version", "name"}).
			AddRow(migration4.Version, time.Now(), migration4.Name),
	)
	mock.ExpectExec(migration4.Discard.Statements[0]).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`INSERT INTO migrations(version, date, name) values(3, NOW(), 'roles_table')`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	mock.ExpectBegin()
	mock.ExpectQuery(versionQuery).WillReturnRows(
		sqlmock.NewRows([]string{"date", "version", "name"}).
			AddRow(migration3.Version, time.Now(), migration3.Name),
	)
	mock.ExpectExec(migration3.Discard.Statements[0]).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`INSERT INTO migrations(version, date, name) values(2, NOW(), 'users_email_index')`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	mock.ExpectBegin()
	mock.ExpectQuery(versionQuery).WillReturnRows(
		sqlmock.NewRows([]string{"date", "version", "name"}).
			AddRow(migration2.Version, time.Now(), migration2.Name),
	)
	mock.ExpectExec(migration2.Discard.Statements[0]).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`INSERT INTO migrations(version, date, name) values(1, NOW(), 'users_table')`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	mock.ExpectBegin()
	mock.ExpectQuery(versionQuery).WillReturnRows(
		sqlmock.NewRows([]string{"date", "version", "name"}).
			AddRow(migration1.Version, time.Now(), migration1.Name),
	)
	mock.ExpectExec(migration1.Discard.Statements[0]).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`INSERT INTO migrations(version, date, name) values(0, NOW(), 'create_migrations_table')`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	mock.ExpectBegin()
	mock.ExpectQuery(versionQuery).WillReturnRows(
		sqlmock.NewRows([]string{"date", "version", "name"}).
			AddRow(migration0.Version, time.Now(), migration0.Name),
	)
	mock.ExpectExec(migration0.Discard.Statements[0]).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	m, err := New(mdb, StdLog, migrations)
	if err != nil {
		t.Fatalf("failed to create migrate: %s", err)
	}

	if err := m.Down(context.Background()); err != nil {
		t.Fatalf("migration run failed: %s", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("mock expectations failed: %s", err)
	}
}
