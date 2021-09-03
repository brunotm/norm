# statement is a simple SQL query builder for Go

![Build Status](https://github.com/brunotm/statement/actions/workflows/test.yml/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/brunotm/statement?cache=0)](https://goreportcard.com/report/brunotm/statement)
[![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white&style=flat-square)](https://pkg.go.dev/github.com/brunotm/statement)
[![Apache 2 licensed](https://img.shields.io/badge/license-Apache2-blue.svg)](https://raw.githubusercontent.com/brunotm/statement/master/LICENSE)

This package provides a simple SQL query builder for Go which aims for simplicity, performance
and to be out of the way when working with SQL code in Go.

**It is not and will not be a magic ORM or DSL that hides the SQL code and logic.**

## Currently supported

	* Select
    	* Comment
    	* Columns
    	* From (table or statement.SelectStatement)
    	* Join
    	* Where
    	* WhereIn
    	* With (statement.SelectStatement)
    	* WithRecursive (statement.SelectStatement)
    	* Having
    	* GroupBy
    	* Order
    	* Limit
    	* Offset
    	* Distinct
    	* ForUpdate
    	* SkipLocked
    	* Union (statement.SelectStatement)
    	* UnionAll (statement.SelectStatement)
	* Insert
    	* Comment
    	* Into
    	* With (statement.SelectStatement)
    	* Returning
    	* Record (from struct)
    	* ValuesSelect (statement.SelectStatement)
    	* OnConflictUpdate
	* Update
    	* Comment
    	* Table
    	* Set
    	* SetMap
    	* With (statement.SelectStatement)
    	* Where
    	* WhereIn
    	* Returning
	* Delete
    	* Comment
    	* From
    	* With (statement.SelectStatement)
    	* Where
    	* WhereIn
    	* Returning
	* DDL
    	* Comment
    	* Create
    	* Alter
    	* Truncate
    	* Drop

## TODO
	* better statement/db tests

## Statement builder Usage (with database/sql)

```go
	type Part struct {
		ID string
		Name string
	}

	part := Part{
		Name: "part01"
	}

	query, err := statement.Select().WithRecursive(
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

## statement/db

The statement/db package provides a thin wrapper around a `sql.DB` which enforces transactional
access to the database with configurable isolation levels, transaction scoped query cache and scanning `sql.Rows` into structs and query logging.

```go
	conf := db.Config{
		Log: func(message, id string, err error, d time.Duration, query string) {
			log.Println(
				"message:", message,
				", id:", id,
				", error:", err,
				", duration_millis:" d.Milliseconds(),
				", query": query)
		},
		ReadOpt:  sql.LevelSerializable,
		WriteOpt: sql.LevelSerializable,
	}

	d, err := db.New(sdb, conf)
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

	stmt := statement.Select().WithRecursive(
	"included_parts",
	Select().Columns("sub_part", "part", "quantity").
		From("parts").Where("part = ?", part.Name).
		UnionAll(
			Select().Columns("p.sub_part", "p.part", "p.quantity").
				From("included_parts AS pr").
				JoinInner("parts AS p", "p.part = pr.sub_part"),
		),
	).Columns("sub_part as name", "SUM(quantity) as total_quantity").
	From("included_parts").GroupBy("name")

	var parts []Part
	err = tx.Query(&parts, stmt)
	if err != nil {
	// handle err
	}

	// do things with parts
```

## Install

This package has no external dependencies and requires Go >= 1.16 and modules.

```shell
go get -u -v github.com/brunotm/statement
```

## Test

```shell
go test -v -cover .
```

## Acknowledgements
The statement package is inspired by the work done on [mailru/dbr](https://github.com/mailru/dbr)
