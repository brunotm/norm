# norm/migrate

Migrations can be provided as:
* A set of files in a `fs.FS` (which can be embedded) matching the `(\d+)_(\w+)\.(apply|discard)\.sql` naming pattern using `migrate.NewWithFiles()`
* A `[]*migrate.Migration using` using `migrate.New()`

By default migrations can have multiple SQL statements and are run within database transactions. Transactions can be disabled, limiting each migration to single SQL statement.

## Using migration files
Each migration file can contain multiple SQL statements and each individual statement must be terminated with `;`.

To disable transactions for a given migration annotate the migration file with the following SQL comment `-- migrate: NoTransaction`.

### Example

**Migration files structure**

```sh
-- versions
	|-- 0001_orders_table.apply.sql
	|-- 0001_orders_table.discard.sql
	|-- 0002_users_table.apply.sql
	|-- 0002_users_table.discard.sql
	|-- 0003_payments_table.apply.sql
	|-- 0003_payments_table.discard.sql
```

**Migration code**

```go
	ctx := context.Background()

	db, err := sql.Open("pgx", "postgres://user:pass@postgres:5432/db")
	if err != nil {
		// handle err
	}

	// use a local dir with migrations, could also be a embed.FS
	//go:embed versions/*
	// var versions embed.FS
	m, err := migrate.NewWithFiles(db, log.Printf, os.DirFS("./versions"))
	if err != nil {
		// handle err
	}

	// migrate all the way up
	err = m.Up(ctx)
	if err != nil {
		// handle err
	}

	// get current version
	v, err := m.Version(ctx)
	if err != nil {
		// handle err
	}

	fmt.Println(v.Version)


	// migrate forward or backward to a specific version
	err = m.Apply(ctx, 2)
	if err != nil {
		panic(err)
	}

	// migrate all the way down and remove migration history
	err = m.Down(ctx)
	if err != nil {
		panic(err)
	}
```

## Using migration structs
Each migration struct can contain multiple SQL statements and each individual statement must NOT be terminated with `;`.

To disable transactions for a given migration, set the `migrate.Migration.NoTx` to `true`.

### Example

**Migration structs**

```go

var versions = []*Migration{
	{
		Version: 1,
		Name:    "users_table",
		Apply: Statements{
			NoTx:       true,
			Statements: []string{"CREATE TABLE IF NOT EXISTS users(id text, name text, email text, role text, PRIMARY KEY (id))"},
		},
		Discard: Statements{
			Statements: []string{"DROP TABLE IF EXISTS users CASCADE"},
		},
	},
	{
		Version: 2,
		Name:    "users_email_index",
		Apply: Statements{
			Statements: []string{"CREATE INDEX IF NOT EXISTS ix_users_email ON users (email)"},
		},
		Discard: Statements{
			Statements: []string{"DROP INDEX IF EXISTS ix_users_email CASCADE"},
		},
	},
	{
		Version: 3,
		Name:    "roles_table",
		Apply: Statements{
			Statements: []string{"CREATE TABLE IF NOT EXISTS roles(id text, name text, properties jsonb NOT NULL DEFAULT '{}'::jsonb, PRIMARY KEY (id))"},
		},
		Discard: Statements{
			Statements: []string{"DROP TABLE IF EXISTS roles CASCADE"},
		},
	},
	{
		Version: 4,
		Name:    "user_roles_fk",
		Apply: Statements{
			Statements: []string{"ALTER TABLE users ADD CONSTRAINT roles_fk FOREIGN KEY (role) REFERENCES roles (id)"},
		},
		Discard: Statements{
			Statements: []string{"ALTER TABLE users DROP CONSTRAINT roles_fk CASCADE"},
		},
	},
}

```

**Migration code**

```go
	ctx := context.Background()

	db, err := sql.Open("pgx", "postgres://user:pass@postgres:5432/db")
	if err != nil {
		// handle err
	}

	// use a local dir with migrations, could also be a embed.FS
	//go:embed versions/*
	// var versions embed.FS
	m, err := migrate.New(db, log.Printf, versions)
	if err != nil {
		// handle err
	}

	// migrate all the way up
	err = m.Up(ctx)
	if err != nil {
		// handle err
	}

	// get current version
	v, err := m.Version(ctx)
	if err != nil {
		// handle err
	}

	fmt.Println(v.Version)


	// migrate forward or backward to a specific version
	err = m.Apply(ctx, 2)
	if err != nil {
		panic(err)
	}

	// migrate all the way down and remove migration history
	err = m.Down(ctx)
	if err != nil {
		panic(err)
	}
```