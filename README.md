# statement is a simple SQL query builder for Go

![Build Status](https://github.com/brunotm/statement/actions/workflows/test.yml/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/brunotm/statement?cache=0)](https://goreportcard.com/report/brunotm/statement)
[![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white&style=flat-square)](https://pkg.go.dev/github.com/brunotm/statement)
[![Apache 2 licensed](https://img.shields.io/badge/license-Apache2-blue.svg)](https://raw.githubusercontent.com/brunotm/statement/master/LICENSE)

This package provides a simple SQL query builder for Go which aims for simplicity, performance and to
stand out of the way when building complex or simple queries.

The motivation for this package is to allow expressiveness, flexibility and performance without cruft and overhead when handling SQL queries and code in Go.


It does not intend to be used as an ORM or a DSL for building SQL statements, but rather provide a simple and flexible structure for working with SQL code in Go.


## Install

This package requires Go modules.

```shell
go get github.com/brunotm/statement
```

## Usage

```go
query, err := statement.Select("sub_part", "SUM(quantity) as total_quantity").
	From("included_parts").GroupBy("sub_part").
	WithRecursive(
		"included_parts",
		Select("sub_part", "part", "quantity").
			From("parts").Where("part = ?", "our_product").
			UnionAll(
				Select("p.sub_part", "p.part", "p.quantity").
					From("included_parts AS pr").
					JoinInner("parts AS p", "p.part = pr.sub_part"),
			),
	).String()

	if err != nil {
		// handle error
	}

	rows, err := tx.Query(query)
	// handle rows.....
```

## Test

```shell
go test -v -cover .
```

## Acknowledgements
The statement package is inspired by the great [mailru/dbr](https://github.com/mailru/dbr)