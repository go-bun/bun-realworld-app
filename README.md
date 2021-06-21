# Go treemux + bun realworld application

[![build workflow](https://github.com/go-bun/bun-realworld-app/actions/workflows/build.yml/badge.svg)](https://github.com/go-bun/bun-realworld-app/actions)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/uptrace/bun-realworld-app)](https://pkg.go.dev/github.com/uptrace/bun-realworld-app)

## Introduction

This project implements RealWorld JSON API as specified in the
[spec](https://github.com/gothinkster/realworld). It was created to demonstrate how to use:

- [treemux HTTP router](https://github.com/vmihailenco/treemux).
- [Bun DB](https://github.com/uptrace/bun).
- [bun/migrate](https://bun.uptrace.dev/guide/migrations.html).
- [bun/dbfixture](https://bun.uptrace.dev/guide/fixtures.html).

## Project structure

The project uses Bun [starter kit](https://bun.uptrace.dev/guide/starter-kit.html) and consists of
the following packages:

- [bunapp](bunapp) package parses configs, establishes DB connections etc.
- [org](org) package manages users and tokens.
- [blog](blog) package manages articles and comments.
- [cmd/bun](cmd/bun) provides CLI commands to run HTTP server and work with DB.
- [cmd/bun/migrations](cmd/bun/migrations) contains database migrations.

The most interesting part for Bun users is probably [article filter](blog/article_filter.go).

## Project bootstrap

Project comes with a `Makefile` that contains following recipes:

- `make db_reset` drops existing database and creates a new one.
- `make test` runs unit tests.
- `make api_test` runs API tests provided by
  [RealWorld](https://github.com/gothinkster/realworld/tree/master/api).

After checking that tests are passing you can run HTTP server:

```shell
go run cmd/bun/main.go -env=dev runserver
```
