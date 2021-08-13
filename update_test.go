package statement

import (
	"testing"
)

var (
	updateCases = []struct {
		name    string
		expect  string
		stmt    Statement
		wantErr bool
	}{
		{
			name:   "simple_set",
			expect: `UPDATE users SET email = 'john.doe@email.com', role = 'admin', user = 'john.doe' WHERE id = 123`,
			stmt: Update("users").Set("user", "john.doe").Set("email", "john.doe@email.com").
				Set("role", "admin").Where("id = ?", 123),
			wantErr: false,
		},
		{
			name:   "simple_setmap",
			expect: `UPDATE users SET email = 'john.doe@email.com', role = 'admin', user = 'john.doe' WHERE id = 123`,
			stmt: Update("users").SetMap(map[string]interface{}{
				"user":  "john.doe",
				"email": "john.doe@email.com",
				"role":  "admin",
			}).Where("id = ?", 123),
			wantErr: false,
		},
		{
			name:   "where_in",
			expect: `UPDATE users SET email = 'john.doe@email.com', role = 'admin', user = 'john.doe' WHERE id IN (123,321)`,
			stmt: Update("users").SetMap(map[string]interface{}{
				"user":  "john.doe",
				"email": "john.doe@email.com",
				"role":  "admin",
			}).WhereIn("id", 123, 321),
			wantErr: false,
		},
		{
			name:   "with",
			expect: `WITH select_offices AS (SELECT country,city,address,postal_code FROM offices WHERE country IN ('uk','es','pt','fr')) UPDATE users SET email = 'john.doe@email.com', role = 'admin', user = 'john.doe' WHERE id IN (123,321)`,
			stmt: Update("users").
				With("select_offices", Select("country", "city", "address", "postal_code").
					From("offices").WhereIn("country", "uk", "es", "pt", "fr")).
				SetMap(map[string]interface{}{
					"user":  "john.doe",
					"email": "john.doe@email.com",
					"role":  "admin",
				}).WhereIn("id", 123, 321),
			wantErr: false,
		},
		{
			name:   "returning",
			expect: `UPDATE users SET email = 'john.doe@email.com', role = 'admin', user = 'john.doe' WHERE id IN (123,321) RETURNING email`,
			stmt: Update("users").SetMap(map[string]interface{}{
				"user":  "john.doe",
				"email": "john.doe@email.com",
				"role":  "admin",
			}).WhereIn("id", 123, 321).Returning("email"),
			wantErr: false,
		},
	}
)

func TestUpdate(t *testing.T) {
	for _, tt := range updateCases {
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
