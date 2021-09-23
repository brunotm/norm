# norm (NoORM)

[![Build Status](https://github.com/brunotm/norm/actions/workflows/test.yml/badge.svg)](https://github.com/brunotm/norm/actions)
[![Go Report Card](https://goreportcard.com/badge/brunotm/norm?cache=0)](https://goreportcard.com/report/brunotm/norm)
[![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white&style=flat-square)](https://pkg.go.dev/github.com/brunotm/norm)
[![Apache 2 licensed](https://img.shields.io/badge/license-Apache2-blue.svg)](https://raw.githubusercontent.com/brunotm/norm/master/LICENSE)


This repository provides a **no frills, no bloat and no magic** collection of utilities for working with SQL code and databases
in Go. It **relies solely on standard library packages** and focuses on simplicity, stability and performance.

# Packages

## [norm/statement](statement/README.md)

A simple and performant SQL query builder for Go which doesn't hide or obscures SQL code and
logic but rather makes them explicit and a fundamental part of the application code.

It handles the all parameter interpolation at the package level using the `?` placeholder, so
queries are mostly portable across databases.

### Features

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
		* OnConflict
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


## [norm/database](database/README.md)

A safe sql.DB wrapper which enforces transactional access to the database, transaction query caching and operation logging and plays nicely with `norm/statement`.

### Features

	* Contextual operation logging
	* Transactional access with default isolation level
	* Cursor for traversing large result sets
	* Row scanning into structs or []struct
	* Transaction scoped query caching
	* Transaction ids for request tracing

## [norm/migrate](migrate/README.md)

A simple database migration package which does the necessary, not less, not more.

### Features

	* Migration sequence management
	* Migrate Up/Down/Apply(<version>)
	* Apply/discard migrations
	* Transactional apply/discard migrations

## Motivation

I like simple, efficient and clean code. There is too much ORM stuff floating around these days.

### Install
```shell
go get -u -v github.com/brunotm/norm
```

## Test

```shell
go test -v -cover ./...
```
