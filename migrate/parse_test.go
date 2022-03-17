package migrate

import (
	"reflect"
	"testing"
)

func TestParseSimple(t *testing.T) {
	stmt, err := parseStatement(stmt)
	if err != nil {
		t.Fatalf("failed to parse statement: %s", err)
	}

	if !reflect.DeepEqual(expected, stmt) {
		t.Fatalf("expected: %#v got: %#v", expected, stmt)
	}
}

func TestParseMultiNoTx(t *testing.T) {
	notx := append([]byte(`-- migrate: NoTransaction`), stmt...)
	_, err := parseStatement(notx)

	if err != ErrInvalidNoTx {
		t.Fatalf("failed to parse statement: %s", err)
	}
}

var stmt = []byte(`
CREATE TABLE IF NOT EXISTS users (
	created_at timestamptz NOT NULL DEFAULT now(),
	updated_at timestamptz NOT NULL DEFAULT now(),
	id UUID,
	name text NOT NULL,
	email text NOT NULL,
	PRIMARY KEY (id)
);
CREATE UNIQUE INDEX IF NOT EXISTS ix_unique_users_name ON users (name);
CREATE UNIQUE INDEX IF NOT EXISTS ix_unique_users_email ON users (email);
CREATE INDEX IF NOT EXISTS ix_users_created_at ON users (created_at);
CREATE INDEX IF NOT EXISTS ix_users_updated_at ON users (updated_at);

`)

var expected = Statements{
	NoTx: false,
	Statements: []string{
		"CREATE TABLE IF NOT EXISTS users ( created_at timestamptz NOT NULL DEFAULT now(), updated_at timestamptz NOT NULL DEFAULT now(), id UUID, name text NOT NULL, email text NOT NULL, PRIMARY KEY (id) )",
		"CREATE UNIQUE INDEX IF NOT EXISTS ix_unique_users_name ON users (name)",
		"CREATE UNIQUE INDEX IF NOT EXISTS ix_unique_users_email ON users (email)",
		"CREATE INDEX IF NOT EXISTS ix_users_created_at ON users (created_at)",
		"CREATE INDEX IF NOT EXISTS ix_users_updated_at ON users (updated_at)",
	},
}
