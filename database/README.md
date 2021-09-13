# norm/database

## Example

```go
	db, err := database.New(sdb, sql.LevelSerializable, database.DefaultLogger)
	if err != nil {
    	// handle err
	}

	type Part struct {
    	ID string
    	Name string
    	TotalQuantity int64
	}

	part := Part{
		Name: "part01"
	}

	tx, err := db.Read(ctx, id)
	if err != nil {
		// handle err
	}

	query, err := statement.Select().Comment("request-id: ?", id).
		WithRecursive(
			"included_parts",
			Select().Columns("sub_part", "part", "quantity").
				From("parts").Where("part = ?", part.Name).
				UnionAll(
					Select().Columns("p.sub_part", "p.part", "p.quantity").
						From("included_parts AS pr").
						JoinInner("parts AS p", "p.part = pr.sub_part"),
				),
		).Columns("sub_part", "SUM(quantity) as total_quantity").
			From("included_parts").GroupBy("sub_part").String()

	var parts []Part
	if err = tx.Query(&parts, stmt); err != nil {
	// handle err
	}

	// do things with parts
```