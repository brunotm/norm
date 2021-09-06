# norm/migrate

The migration files in the provided `fs.FS` must be on its root (no directory nesting) and use the `(\d+)_(\w+)\.(apply|discard)\.sql` naming pattern.

Example:
```sh
.
`-- versions
	|-- 0001_orders_table.apply.sql
	|-- 0001_orders_table.discard.sql
	|-- 0002_users_table.apply.sql
	|-- 0002_users_table.discard.sql
	|-- 0003_payments_table.apply.sql
	`-- 0003_payments_table.discard.sql
```

In the current implementation each migration for either apply or discard is run inside its own transaction.

## TODO
* Add support for running migrations outside a transaction

## Example

```go
	ctx := context.Background()

	db, err := sql.Open("pgx", "postgres://user:pass@postgres:5432/db")
	if err != nil {
		// handle err
	}

	// use a local dir with migrations, could also be a embed.FS
	//go:embed versions/*
	// var versions embed.FS
	m, err := migrate.NewWithFiles(db, os.DirFS("./versions"), log.Printf)
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