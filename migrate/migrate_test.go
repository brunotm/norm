package migrate

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestMigrateVersions(t *testing.T) {
	mdb, _, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("error opening mock database: %s", err)
	}
	defer mdb.Close()

	m, err := New(mdb, StdLog, migrations)
	if err != nil {
		t.Fatalf("failed to create migrate: %s", err)
	}

	if versions := m.Versions(); len(versions) != len(migrations)+1 {
		t.Fatalf("wrong version count: %d, expected: %d, data %#v", len(versions), len(migrations)+1, versions)
	}
}
