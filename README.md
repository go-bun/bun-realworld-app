# Go treemux + bun realworld application

[![build workflow](https://github.com/go-bun/bun-realworld-app/actions/workflows/build.yml/badge.svg)](https://github.com/go-bun/bun-realworld-app/actions)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/uptrace/bun-realworld-app)](https://pkg.go.dev/github.com/uptrace/bun-realworld-app)

## Introduction

This project implements RealWorld JSON API as specified in the
[spec](https://github.com/gothinkster/realworld). It was created to demonstrate how to use:

- [treemux HTTP router](https://github.com/vmihailenco/treemux).
- [bun db](https://github.com/uptrace/bun).
- [bun migrations](https://bun.uptrace.dev/guide/migrations.html).
- [bun fixtures](https://bun.uptrace.dev/guide/fixtures.html).

## Project structure

The project consists of the following packages:

- [app](app) package parses configs, establishes DB connections etc.
- [org](org) package manages users and tokens.
- [blog](blog) package manages articles and comments.
- [cmd/bun](cmd/bun) provides CLI commands to run HTTP server and work with DB.
- [cmd/bun/migrations](cmd/bun/migrations) contains database migrations.

The most interesting part for bun users is probably [article filter](blog/article_filter.go).

## Project bootstrap

First of all you need to create a config file changing defaults as needed:

```
cp app/config/dev.yaml.default app/config/dev.yaml
```

Project comes with a `Makefile` that contains following recipes:

- `make db_reset` drops existing database and creates a new one.
- `make test` runs unit tests.
- `make api_test` runs API tests provided by
  [RealWorld](https://github.com/gothinkster/realworld/tree/master/api).

After checking that tests are passing you can run HTTP server:

```shell
go run cmd/bun/*.go -env=dev runserver
```
