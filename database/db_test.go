package database

import (
	"context"
	"database/sql"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/brunotm/norm/statement"
)

func TestTxExecSimple(t *testing.T) {
	mdb, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("error opening mock database: %s", err)
	}
	defer mdb.Close()

	db, err := New(mdb, sql.LevelSerializable, DefaultLogger)
	if err != nil {
		t.Fatalf("error opening norm/database.DB: %s", err)
	}

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO users(id,name,email,role) VALUES ('123abc','john doe','johnd@email.com','admin')").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	tx, err := db.Update(context.Background(), "")
	if err != nil {
		t.Fatalf("error opening norm/database.DB transaction: %s", err)
	}

	insert := statement.Insert().Into("users").Columns("id", "name", "email", "role").
		Values("123abc", "john doe", "johnd@email.com", "admin")

	_, err = tx.Exec(insert)
	if err != nil {
		t.Fatalf("error executing norm/database.DB transaction: %s", err)
	}

	if err = tx.Commit(); err != nil {
		t.Fatalf("error committing norm/database.DB transaction: %s", err)
	}

	if err = mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("mock expectations failed: %s", err)
	}
}

func TestTxQuerySimple(t *testing.T) {
	mdb, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("error opening mock database: %s", err)
	}
	defer mdb.Close()

	db, err := New(mdb, sql.LevelSerializable, DefaultLogger)
	if err != nil {
		t.Fatalf("error opening norm/database.DB: %s", err)
	}

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT id,name,email,role FROM users").WillReturnRows(
		sqlmock.NewRows([]string{"id", "name", "email", "role"}).
			AddRow("123abc", "john doe", "johnd@email.com", "admin").
			AddRow("123abcd", "jane doe", "janed@email.com", "user").
			AddRow("123abcde", "susan vix", "susanv@email.com", "moderator"),
	)
	mock.ExpectRollback()

	tx, err := db.Read(context.Background(), "someid")
	if err != nil {
		t.Fatalf("error opening norm/database.DB transaction: %s", err)
	}

	query := statement.Select().Columns("id", "name", "email", "role").From("users")

	type user struct {
		ID    string
		Name  string
		Email string
		Role  string
	}
	var users []user

	if err = tx.Query(&users, query); err != nil {
		t.Fatalf("error performing norm/database.DB query: %s", err)
	}

	if err = tx.Rollback(); err != nil {
		t.Fatalf("error rolling back transaction: %s", err)
	}

	if len(users) > 3 {
		t.Fatalf("expected 3 rows, got %d, data: %#v", len(users), users)
	}

	if err = mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("mock expectations failed: %s", err)
	}

}

func TestTxQueryCache(t *testing.T) {
	mdb, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("error opening mock database: %s", err)
	}
	defer mdb.Close()

	db, err := New(mdb, sql.LevelSerializable, DefaultLogger)
	if err != nil {
		t.Fatalf("error opening norm/database.DB: %s", err)
	}

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT id,name,email,role FROM users").WillReturnRows(
		sqlmock.NewRows([]string{"id", "name", "email", "role"}).
			AddRow("123abc", "john doe", "johnd@email.com", "admin").
			AddRow("123abcd", "jane doe", "janed@email.com", "user").
			AddRow("123abcde", "susan vix", "susanv@email.com", "moderator"),
	)
	mock.ExpectRollback()

	tx, err := db.Read(context.Background(), "someid")
	if err != nil {
		t.Fatalf("error opening norm/database.DB transaction: %s", err)
	}

	query := statement.Select().Columns("id", "name", "email", "role").From("users")

	type user struct {
		ID    string
		Name  string
		Email string
		Role  string
	}
	var users []user

	if err = tx.QueryCache(&users, query); err != nil {
		t.Fatalf("error performing norm/database.DB query: %s", err)
	}

	// running the query the 2md time should not hit tha database and fail expectations
	if err = tx.QueryCache(&users, query); err != nil {
		t.Fatalf("error performing norm/database.DB query: %s", err)
	}

	if err = tx.Rollback(); err != nil {
		t.Fatalf("error rolling back transaction: %s", err)
	}

	if len(users) > 3 {
		t.Fatalf("expected 3 rows, got %d, data: %#v", len(users), users)
	}

	if err = mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("mock expectations failed: %s", err)
	}

}

func TestDBPing(t *testing.T) {

	mdb, mock, err := sqlmock.New(
		sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual),
		sqlmock.MonitorPingsOption(true),
	)
	if err != nil {
		t.Fatalf("error opening mock database: %s", err)
	}
	defer mdb.Close()

	db, err := New(mdb, sql.LevelSerializable, nil)
	if err != nil {
		t.Fatalf("error opening norm/database.DB: %s", err)
	}

	mock.ExpectPing()
	if err = db.Ping(context.Background()); err != nil {
		t.Fatalf("error pinging the database: %s", err)
	}

	if err = mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("mock expectations failed: %s", err)
	}
}

func TestDBClose(t *testing.T) {
	mdb, mock, err := sqlmock.New(
		sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual),
		sqlmock.MonitorPingsOption(true),
	)
	if err != nil {
		t.Fatalf("error opening mock database: %s", err)
	}
	defer mdb.Close()

	db, err := New(mdb, sql.LevelSerializable, nil)
	if err != nil {
		t.Fatalf("error opening norm/database.DB: %s", err)
	}

	mock.ExpectClose()
	if err = db.Close(); err != nil {
		t.Fatalf("error closing the database: %s", err)
	}

	if err = mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("mock expectations failed: %s", err)
	}
}
