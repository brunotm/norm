package statement

import "testing"

var (
	selectCases = []struct {
		name    string
		expect  string
		stmt    Statement
		wantErr bool
	}{
		{
			name:   "simple",
			expect: `SELECT id,user,email,role FROM users WHERE email = 'john.doe@email.com' AND role IN ('admin','owner')`,
			stmt: Select().Columns("id", "user", "email", "role").From("users").Where("email = ?", "john.doe@email.com").
				WhereIn("role", "admin", "owner"),
			wantErr: false,
		},
		{
			name:   "order_asc",
			expect: `SELECT id,user,email,role FROM users WHERE email = 'john.doe@email.com' AND role IN ('admin','owner') ORDER BY id,email ASC`,
			stmt: Select().Columns("id", "user", "email", "role").From("users").Where("email = ?", "john.doe@email.com").
				WhereIn("role", "admin", "owner").OrderAsc("id", "email"),
			wantErr: false,
		},
		{
			name:   "order_desc",
			expect: `SELECT id,user,email,role FROM users WHERE email = 'john.doe@email.com' AND role IN ('admin','owner') ORDER BY id,email DESC`,
			stmt: Select().Columns("id", "user", "email", "role").From("users").Where("email = ?", "john.doe@email.com").
				WhereIn("role", "admin", "owner").OrderDesc("id", "email"),
			wantErr: false,
		},
		{
			name:   "limit_offset",
			expect: `SELECT id,user,email,role FROM users WHERE email = 'john.doe@email.com' AND role IN ('admin','owner') LIMIT 10 OFFSET 0`,
			stmt: Select().Columns("id", "user", "email", "role").From("users").Where("email = ?", "john.doe@email.com").
				WhereIn("role", "admin", "owner").Limit(10),
			wantErr: false,
		},
		{
			name:   "with",
			expect: `WITH select_offices AS (SELECT country,city,address,postal_code FROM offices WHERE country IN ('uk','es','pt','fr')) SELECT * FROM users INNER JOIN select_offices ON users.city = select_offices.city WHERE email = 'john.doe@email.com' AND role IN ('admin','owner')`,
			stmt: Select().With("select_offices",
				Select().Columns("country", "city", "address", "postal_code").From("offices").WhereIn("country", "uk", "es", "pt", "fr")).
				Columns("*").From("users").JoinInner("select_offices", "users.city = select_offices.city").
				Where("email = ?", "john.doe@email.com").
				WhereIn("role", "admin", "owner"),
			wantErr: false,
		},
		{
			name:   "with_recursive_union",
			expect: `WITH RECURSIVE included_parts AS (SELECT sub_part,part,quantity FROM parts WHERE part = 'our_product' UNION ALL SELECT p.sub_part,p.part,p.quantity FROM included_parts AS pr INNER JOIN parts AS p ON p.part = pr.sub_part) SELECT sub_part,SUM(quantity) as total_quantity FROM included_parts GROUP BY sub_part`,
			stmt: Select().WithRecursive(
				"included_parts",
				Select().Columns("sub_part", "part", "quantity").
					From("parts").Where("part = ?", "our_product").
					UnionAll(
						Select().Columns("p.sub_part", "p.part", "p.quantity").
							From("included_parts AS pr").
							JoinInner("parts AS p", "p.part = pr.sub_part"),
					),
			).Columns("sub_part", "SUM(quantity) as total_quantity").
				From("included_parts").GroupBy("sub_part"),
			wantErr: false,
		},
	}
)

func TestSelect(t *testing.T) {
	for _, tt := range selectCases {
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
