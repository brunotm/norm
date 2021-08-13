package statement

import (
	"testing"
)

var (
	insertCases = []struct {
		name    string
		expect  string
		stmt    Statement
		wantErr bool
	}{
		{
			name:    "simple",
			expect:  `INSERT INTO users(id,user,email,role) VALUES (123,'john.doe','john.doe@email.com','admin')`,
			stmt:    Insert("users").Columns("id", "user", "email", "role").Values(123, "john.doe", "john.doe@email.com", "admin"),
			wantErr: false,
		},
		{
			name:   "from_select",
			expect: `INSERT INTO users(id,user,email,role) (SELECT id,user,email,role FROM old_users INNER JOIN roles ON old_users.id = roles.user_id)`,
			stmt: Insert("users").Columns("id", "user", "email", "role").ValuesSelect(
				Select("id", "user", "email", "role").From("old_users").JoinInner("roles", "old_users.id = roles.user_id")),
			wantErr: false,
		},
		{
			name:   "on_conflict",
			expect: `INSERT INTO users(id,user,email,role) VALUES (123,'john.doe','john.doe@email.com','admin') ON CONFLICT ON CONSTRAINT users_pkey DO UPDATE SET email = 'john.doe@email.com', role = 'admin', user = 'john.doe'`,
			stmt: Insert("users").Columns("id", "user", "email", "role").Values(123, "john.doe", "john.doe@email.com", "admin").
				OnConflictUpdate("users_pkey", map[string]interface{}{
					"user":  "john.doe",
					"email": "john.doe@email.com",
					"role":  "admin",
				}),
			wantErr: false,
		},
		{
			name:    "returning",
			expect:  `INSERT INTO users(id,user,email,role) VALUES (123,'john.doe','john.doe@email.com','admin') RETURNING id`,
			stmt:    Insert("users").Columns("id", "user", "email", "role").Values(123, "john.doe", "john.doe@email.com", "admin").Returning("id"),
			wantErr: false,
		},
		{
			name: "invalid_with_alias",
			stmt: Insert("users").Columns("id", "user", "email", "role").
				Values(123, "john.doe", "john.doe@email.com", "admin").With("", Select("id").From("roles")),
			wantErr: true,
		},
	}
)

func TestInsert(t *testing.T) {
	for _, tt := range insertCases {
		t.Run(tt.name, func(t *testing.T) {
			s, err := tt.stmt.String()
			if !tt.wantErr && err != nil {
				t.Fatalf("error building statement: %s", err)
			}

			if tt.expect != s {
				t.Fatalf("expected: %s, got: %s", tt.expect, s)
			}
		})
	}
}
