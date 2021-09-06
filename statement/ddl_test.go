package statement

import "testing"

var (
	ddlCases = []struct {
		name    string
		expect  string
		stmt    Statement
		wantErr bool
	}{
		{
			name:    "create",
			expect:  `CREATE INDEX IF NOT EXISTS ix_users_created_at ON users (created_at)`,
			stmt:    Create("INDEX IF NOT EXISTS ? ON ? (?)", "ix_users_created_at", "users", "created_at"),
			wantErr: false,
		},
		{
			name:    "alter",
			expect:  `ALTER TABLE users ADD COLUMN address text`,
			stmt:    Alter("TABLE ? ADD COLUMN ? ?", "users", "address", "text"),
			wantErr: false,
		},
		{
			name:    "drop",
			expect:  `DROP TABLE users CASCADE`,
			stmt:    Drop("TABLE ? CASCADE", "users"),
			wantErr: false,
		},
		{
			name:    "truncate",
			expect:  `TRUNCATE TABLE users CASCADE`,
			stmt:    Truncate("TABLE ? CASCADE", "users"),
			wantErr: false,
		},
		{
			name: "comment",
			expect: `-- request id: 12435
TRUNCATE TABLE users CASCADE`,
			stmt:    Truncate("TABLE ? CASCADE", "users").Comment("request id: ?", 12435),
			wantErr: false,
		},
	}
)

func TestDDL(t *testing.T) {
	for _, tt := range ddlCases {
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
