# Fantastic coffee (decaffeinated)

This repository contains the basic structure for [Web and Software Architecture](http://gamificationlab.uniroma1.it/en/wasa/) homework project.
It has been described in class.

"Fantastic coffee (decaffeinated)" is a simplified version for the WASA course, not suitable for a production environment.
The full version can be found in the "Fantastic Coffee" repository.

## Project structure

* `cmd/` contains all executables; Go programs here should only do "executable-stuff", like reading options from the CLI/env, etc.
	* `cmd/healthcheck` is an example of a daemon for checking the health of servers daemons; useful when the hypervisor is not providing HTTP readiness/liveness probes (e.g., Docker engine)
	* `cmd/webapi` contains an example of a web API server daemon
* `demo/` contains a demo config file
* `doc/` contains the documentation (usually, for APIs, this means an OpenAPI file)
* `service/` has all packages for implementing project-specific functionalities
	* `service/api` contains an example of an API server
	* `service/database` contains the data access layer: the `AppDatabase` interface and its SQLite implementation. The schema is managed with SQL migrations (see "Database migrations" below).
	* `service/globaltime` contains a wrapper package for `time.Time` (useful in unit testing)
* `vendor/` is managed by Go, and contains a copy of all dependencies
* `webui/` is an example of a web frontend in Vue.js; it includes:
	* Bootstrap JavaScript framework
	* a customized version of "Bootstrap dashboard" template
	* feather icons as SVG
	* Go code for release embedding

Other project files include:
* `open-node.sh` starts a new (temporary) container using `node:20` image for safe and secure web frontend development (you don't want to use `node` in your system, do you?).

## Go vendoring

This project uses [Go Vendoring](https://go.dev/ref/mod#vendoring). You must use `go mod vendor` after changing some dependency (`go get` or `go mod tidy`) and add all files under `vendor/` directory in your commit.

For more information about vendoring:

* https://go.dev/ref/mod#vendoring
* https://www.ardanlabs.com/blog/2020/04/modules-06-vendoring.html

## Database migrations

The database schema is defined by a set of SQL *migrations* stored in
`service/database/migrations/`. The files are embedded in the `webapi` executable at build time and applied automatically at every startup, before the server accepts requests.

### What is a migration?

A migration is a versioned, append-only change to the database schema. Instead of
editing a `schema.sql` by hand (or letting an ORM guess the schema), you describe
each change once, in SQL, with a progressive number. Every environment (your
laptop, a teammate's laptop, the grading machine, production) starts from an empty
database and applies the same migrations in the same order, so everyone ends up
with exactly the same schema.

Migration files become the single source of truth for the schema and are
versioned together with the rest of the code.

### How Fantastic Coffee (decaffeinated) runs migrations

* Migrations live in `service/database/migrations/` and are named with a
  zero-padded numeric prefix: `0001_init.sql`, `0002_add_users.sql`, ...
* File names are sorted lexicographically, so the prefix defines the execution
  order. Keep it zero-padded and always increasing.
* `service/database/migrations.go` embeds the whole directory with `//go:embed`
  and exposes `database.Migrate(ctx, db)`.
* `cmd/webapi/main.go` calls `database.Migrate` right after opening the
  connection and before creating the API router. If a migration fails, the server
  does not start.
* Applied versions are recorded in the `schema_migrations` table (`version`,
  `applied_at`). On the next startup, already-applied migrations are skipped.
* **Each migration runs inside a single transaction** together with the row that
  records it as applied. Either the whole migration is applied and recorded, or
  nothing is. SQLite's DDL is transactional, so this also holds for schema
  changes.
* Migrations are **forward-only**: there is no automatic "undo". If a migration causes a problem, the usual approach is to fix it with a new migration rather than modifying or reverting the already-applied migration. For example, if 0005_add_status.sql introduces an incorrect schema change, create 0006_fix_status.sql that brings the database into the desired state. If the migration caused data corruption or a destructive change that cannot be safely repaired forward, restore the database from a backup taken before the migration and then apply the migrations again.

### Creating a migration

1. Pick the next number. If the last file is `0002_add_users.sql`, create
   `0003_add_posts.sql`.
2. Write plain SQL in it. Multiple statements are allowed.
3. Restart the server (`go run ./cmd/webapi/`): the migration is applied
   automatically.
4. Commit the new `.sql` file together with the code that uses it.

Example (`0003_add_posts.sql`):

```sql
CREATE TABLE posts (
    id     INTEGER NOT NULL PRIMARY KEY,
    author TEXT    NOT NULL,
    body   TEXT    NOT NULL
);

CREATE INDEX idx_posts_author ON posts(author);
```

Keep each migration focused on one logical change, and make sure the whole set of
migrations still produces a correct schema on an **empty** database.

### Rules (please, read twice)

**NEVER modify an existing migration.** Once a migration has been applied
anywhere, its content is frozen. If you edit it, databases that already applied it
will not see your change, while fresh databases will — and the schemas will
silently diverge. This is the most common and most painful mistake.

* **NEVER edit, rename, or delete a migration that has already been applied.**
* **NEVER reuse or reorder a version number.** Two `0003_...` files, or inserting a
  `0002` after `0003` was applied, breaks the ordering guarantees.
* **NEVER make an old migration depend on new code or data.**
* **NEVER put `PRAGMA` statements in a migration**: pragmas cannot run inside a
  transaction. Configure them when opening the connection instead.
* **NEVER commit a migration you have not tested from an empty database.**

Need to change something you already migrated? Write a **new** migration that
changes the schema forward (add a column, drop a table, backfill data, ...). For
example, to rename a column in SQLite: add the new column, copy the data, then
drop the old one — in new migration files.

### Resetting your development database

Since migrations only go forward, the easiest way to start over while developing
is to delete the database file (by default `/tmp/decaf.db`, or whatever
`--db-filename` points to) and restart the server. The full set of migrations is
re-applied to a fresh database.

### Advanced: using an ORM or sqlc

> This is meant for advanced students only. In the course you are expected to use
> the migration system described above. But if you are already an experienced backend developer and want to use advanced tools, keep reading

The migration system is the default and recommended data layer. If you already
know what you are doing and want a different one, the project is designed so you
can plug it in without touching the API handlers:

* All data access goes through the `database.AppDatabase` interface
  (`service/database/database.go`). Handlers depend on that interface, not on the
  concrete SQLite implementation.
* **sqlc**: point its `schema` setting at `service/database/migrations` (keep them
  forward-only, with no "down" files) and its `queries` setting at a directory of
  your own `.sql` files. Commit and vendor the generated code, then write a small
  adapter type that implements `AppDatabase` and pass it to `api.New` in
  `cmd/webapi/main.go`.
* **GORM**: keep running the SQL migrations first and **do not** enable
  `AutoMigrate`: the migrations are the only source of truth for the schema. Map
  your models to the migrated tables and implement `AppDatabase` as a thin
  wrapper.
* In both cases keep `Migrate` in the startup path. The handlers and the `api`
  package should not need any change.

This section is intentionally just a pointer in the right direction: if you take
this route, you are expected to work out the details yourself.

However, keep in mind that any significant changes to the Fantastic Coffee project structure or template are made at your own risk. If such changes introduce problems or incompatibilities, students are responsible for resolving them, and we cannot provide support for issues caused by substantial modifications to the provided template.

## Node/YARN vendoring

This repository uses `yarn` and a vendoring technique that exploits the ["Offline mirror"](https://yarnpkg.com/features/caching). As for the Go vendoring, the dependencies are inside the repository.

You should commit the files inside the `.yarn` directory.

## How to set up a new project from this template

You need to:

* Change the Go module path to your module path in `go.mod`, `go.sum`, and in `*.go` files around the project
* Rewrite the API documentation `doc/api.yaml`
* If no web frontend is expected, remove `webui` and `cmd/webapi/register-webui.go`
* Update top/package comment inside `cmd/webapi/main.go` to reflect the actual project usage, goal, and general info
* Update the code in `run()` function (`cmd/webapi/main.go`) to connect to databases or external resources
* Write API code inside `service/api`, and create any further package inside `service/` (or subdirectories)

## How to build

If you're not using the WebUI, or if you don't want to embed the WebUI into the final executable, then:

```shell
go build ./cmd/webapi/
```

If you're using the WebUI and you want to embed it into the final executable:

```shell
./open-node.sh
# (here you're inside the container)
yarn run build-embed
exit
# (outside the container)
go build -tags webui ./cmd/webapi/
```

## How to run (in development mode)

You can launch the backend only using:

```shell
go run ./cmd/webapi/
```

If you want to launch the WebUI, open a new tab and launch:

```shell
./open-node.sh
# (here you're inside the container)
yarn run dev
```

## How to build for production / homework delivery

```shell
./open-node.sh
# (here you're inside the container)
yarn run build-prod
```

For "Web and Software Architecture" students: before committing and pushing your work for grading, please read the section below named "My build works when I use `yarn run dev`, however there is a Javascript crash in production/grading"

## Known issues

### My build works when I use `yarn run dev`, however there is a Javascript crash in production/grading

Some errors in the code are somehow not shown in `vite` development mode. To preview the code that will be used in production/grading settings, use the following commands:

```shell
./open-node.sh
# (here you're inside the container)
yarn run build-prod
yarn run preview
```

## License

See [LICENSE](LICENSE).
