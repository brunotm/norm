# norm/statement

## Example

```go
	type Part struct {
		ID string
		Name string
	}

	part := Part{
		Name: "part01"
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

	if err != nil {
		// handle error
	}

	rows, err := tx.Query(query)
	// handle rows.....
```
