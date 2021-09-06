package statement

import (
	"testing"
)

var (
	deleteCases = []struct {
		name    string
		expect  string
		stmt    Statement
		wantErr bool
	}{
		{
			name:    "simple",
			expect:  `DELETE FROM users WHERE email = 'john.doe@email.com' AND role = 'admin'`,
			stmt:    Delete().From("users").Where("email = ?", "john.doe@email.com").Where("role = ?", "admin"),
			wantErr: false,
		},
		{
			name:    "where_in",
			expect:  `DELETE FROM users WHERE role IN ('admin','owner')`,
			stmt:    Delete().From("users").WhereIn("role", "admin", "owner"),
			wantErr: false,
		},
		{
			name:   "with",
			expect: `WITH roles_to_delete AS (SELECT id,name FROM roles WHERE expires_at < now()-'1m'::interval) DELETE FROM users WHERE role IN ((SELECT name FROM roles_to_delete))`,
			stmt: Delete().With("roles_to_delete", Select().Columns("id", "name").From("roles").Where("expires_at < now()-?::interval", "1m")).
				From("users").WhereIn("role", Select().Columns("name").From("roles_to_delete")),
			wantErr: false,
		},
		{
			name:    "returning",
			expect:  `DELETE FROM users WHERE email = 'john.doe@email.com' AND role = 'admin' RETURNING id`,
			stmt:    Delete().From("users").Where("email = ?", "john.doe@email.com").Where("role = ?", "admin").Returning("id"),
			wantErr: false,
		},
		{
			name:    "invalid_number_of_arguments",
			stmt:    Delete().From("users").Where("email = ?").Where("role = ?", "admin").Returning("id"),
			wantErr: true,
		},
		{
			name: "comment",
			expect: `-- request id: 12435
DELETE FROM users WHERE email = 'john.doe@email.com' AND role = 'admin' RETURNING id`,
			stmt:    Delete().Comment("request id: ?", 12435).From("users").Where("email = ?", "john.doe@email.com").Where("role = ?", "admin").Returning("id"),
			wantErr: false,
		},
	}
)

func TestDelete(t *testing.T) {
	for _, tt := range deleteCases {
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
