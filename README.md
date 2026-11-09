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
	* `service/api` contains an example of an API server (see "HTTP router" below)
	* `service/database` contains the data access layer: the `AppDatabase` interface and the shared migration runner. Two implementations are included; you can keep both or only the one you need (see "Choosing a database" below):
		* `service/database/sqlite` implements `AppDatabase` with SQLite
		* `service/database/postgres` implements `AppDatabase` with PostgreSQL
	
	  The schema is managed with SQL migrations (see "Database migrations" below).
	* `service/objectstorage` contains the (optional) file storage layer: the `ObjectStorage` interface, and its implementation for S3-compatible services in `service/objectstorage/s3` (see "Object storage" below)
	* `service/globaltime` contains a wrapper package for `time.Time` (useful in unit testing)
* `vendor/` is managed by Go, and contains a copy of all dependencies
* `webui/` is an example of a web frontend in Vue.js; it includes:
	* Bootstrap 5 and Bootstrap Icons (installed with yarn)
	* a small dashboard layout (responsive sidebar, light/dark color mode)
	* Go code for release embedding

Other project files include:
* `open-node.sh` starts a new (temporary) container using `node:24` image for safe and secure web frontend development. It is the same node version used by us during the evaluation of your app.

## HTTP router

A *router* (or *mux*) decides which handler function runs for an incoming HTTP
request, based on its method and path, and extracts the variable parts of the
path (for example the `id` in `/users/{id}`). Go's standard library includes one
(`net/http.ServeMux`), but it is deliberately minimal. This project uses
[chi](https://github.com/go-chi/chi), a small router that is fully compatible
with `net/http` (the HTTP package of the Go standard library): handlers and
middleware are plain
`func(http.ResponseWriter, *http.Request)` functions, so no framework-specific
types are involved and the standard library stays usable (the examples in the
WASA book use plain `net/http`, and the same code works here). chi adds conveniences
the standard router does not provide directly, such as route groups and a
middleware ecosystem. The routes and middleware are wired in `service/api`.

## Go vendoring

This project uses [Go Vendoring](https://go.dev/ref/mod#vendoring). You must use `go mod vendor` after changing some dependency (`go get` or `go mod tidy`) and add all files under `vendor/` directory in your commit.

For more information about vendoring:

* https://go.dev/ref/mod#vendoring
* https://www.ardanlabs.com/blog/2020/04/modules-06-vendoring.html

## Choosing a database

The project includes two complete examples of the data layer, one for each supported database:

| | SQLite | PostgreSQL |
|---|---|---|
| Package | `service/database/sqlite` | `service/database/postgres` |
| What you need | nothing: the database is a local file | a running PostgreSQL server |
| Go driver | `github.com/mattn/go-sqlite3` | `github.com/jackc/pgx/v5` |
| Query placeholders | `?` | `$1`, `$2`, ... |

Both packages implement the same `database.AppDatabase` interface (`service/database/database.go`), so the API handlers
work with either of them without any change.

You can keep both and let whoever runs the program pick one at startup with the configuration (see "Selecting the
database"), or support only one and delete the other (see "Deleting the database you do not use").

Keep in mind that supporting both means doing everything twice: every `AppDatabase` method has to be implemented in
both packages, and every migration has to be written in both `migrations/` directories, with the same file name and the
SQL of each database.

If you are a WASA student: keep in mind that WASAText strictly requires SQLite.

### Selecting the database

The database is selected with the `DB.Driver` configuration option (`sqlite` by default). Like every other option, it
can be set with an environment variable, a command line flag or the configuration file:

| Option | Environment variable | Flag | Used by |
|---|---|---|---|
| `DB.Driver` | `CFG_DB_DRIVER` | `--db-driver` | both (`sqlite` or `postgres`) |
| `DB.Filename` | `CFG_DB_FILENAME` | `--db-filename` | SQLite: path of the database file (default `/tmp/decaf.db`) |
| `DB.DSN` | `CFG_DB_DSN` | `--db-dsn` | PostgreSQL: connection string |

For example, to run with PostgreSQL:

```shell
CFG_DB_DRIVER=postgres CFG_DB_DSN='postgres://user:password@localhost:5432/decaf?sslmode=disable' go run ./cmd/webapi/
```

or, in the configuration file:

```yaml
db:
  driver: postgres
  dsn: postgres://user:password@localhost:5432/decaf?sslmode=disable
```

### Deleting the database you do not use

If you only require to support one database, you can safely delete the other implementation:

1. Delete the directory of the database you do not use: `service/database/postgres/` if you need SQLite, or
   `service/database/sqlite/` if you need PostgreSQL.
2. In `cmd/webapi/main.go`, replace the `switch` on `cfg.DB.Driver` with a single call to the package you kept, and
   remove the import of the deleted package (and of `strings`, which is used only by the `switch`). For example, if you
   kept SQLite, the code becomes:

   ```go
   var db database.AppDatabase
   logger.Infof("using SQLite database %s", cfg.DB.Filename)
   db, err = sqlite.New(context.Background(), cfg.DB.Filename)
   if err != nil {
   	// ... unchanged
   ```
3. Remove the options you no longer use from `cmd/webapi/load-configuration.go` (`Driver`, and either
   `Filename` or `DSN`).
4. Run `go mod tidy` and then `go mod vendor` to remove the unused driver from `go.mod` and `vendor/` (see "Go vendoring"
   above), and check that `go build ./...` still works.

### Differences to keep in mind

The two databases speak slightly different SQL. Write your queries and migrations in the SQL of each database you support:

* **Placeholders**: SQLite uses `?`, PostgreSQL uses `$1`, `$2`, ...
* **Auto-increment IDs**: SQLite uses `INTEGER PRIMARY KEY` (optionally `AUTOINCREMENT`), PostgreSQL uses
  `INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY`.
* **Getting the ID of an inserted row**: with PostgreSQL, `result.LastInsertId()` is **not** supported and always
  returns an error. Use `INSERT ... RETURNING id` with `QueryRow(...).Scan(&id)` instead (this also works with SQLite).
* **Types**: for example binary data is `BLOB` in SQLite and `BYTEA` in PostgreSQL.

## Database migrations

The database schema is defined by a set of SQL *migrations* stored in the `migrations/` directory of the database
package you use (`service/database/sqlite/migrations/` or `service/database/postgres/migrations/`). The files are
embedded in the `webapi` executable at build time and applied automatically at every startup, before the server accepts
requests.

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

* Migrations live in the `migrations/` directory of the database package
  (`service/database/sqlite/migrations/` or `service/database/postgres/migrations/`) and are named with a
  zero-padded numeric prefix: `0001_init.sql`, `0002_add_users.sql`, ...
* File names are sorted lexicographically, so the prefix defines the execution
  order. Keep it zero-padded and always increasing.
* Each database package embeds its `migrations/` directory with `//go:embed`. Its
  `New` function (`sqlite.New` or `postgres.New`) opens the connection and calls
  `database.Migrate` (`service/database/migrations.go`), the migration runner
  shared by both packages.
* `cmd/webapi/main.go` calls `New` before creating the API router. If a migration
  fails, the server does not start.
* Applied versions are recorded in the `schema_migrations` table (`version`,
  `applied_at`). On the next startup, already-applied migrations are skipped.
* **Each migration runs inside a single transaction** together with the row that
  records it as applied. Either the whole migration is applied and recorded, or
  nothing is. Both SQLite and PostgreSQL support transactional DDL, so this also
  holds for schema changes.
* Migrations are **forward-only**: there is no automatic "undo". If a migration causes a problem, the usual approach is to fix it with a new migration rather than modifying or reverting the already-applied migration. For example, if 0005_add_status.sql introduces an incorrect schema change, create 0006_fix_status.sql that brings the database into the desired state. If the migration caused data corruption or a destructive change that cannot be safely repaired forward, restore the database from a backup taken before the migration and then apply the migrations again.

### Creating a migration

1. Pick the next number. If the last file is `0002_add_users.sql`, create
   `0003_add_posts.sql`.
2. Write plain SQL in it. Multiple statements are allowed.
3. Restart the server (`go run ./cmd/webapi/`): the migration is applied
   automatically.
4. Commit the new `.sql` file together with the code that uses it.

Example (`0003_add_posts.sql`) for SQLite:

```sql
CREATE TABLE posts (
    id     INTEGER NOT NULL PRIMARY KEY,
    author TEXT    NOT NULL,
    body   TEXT    NOT NULL
);

CREATE INDEX idx_posts_author ON posts(author);
```

The same migration for PostgreSQL, with an automatically generated `id`:

```sql
CREATE TABLE posts (
    id     INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    author TEXT NOT NULL,
    body   TEXT NOT NULL
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
* **NEVER write idempotent DDL** (for example `CREATE TABLE IF NOT EXISTS`).
  Each migration is applied exactly once; see "Why migrations are not
  idempotent" below.
* For SQLite, **NEVER put `PRAGMA` statements in a migration**: pragmas cannot run
  inside a transaction. Configure them when opening the connection instead.
* **NEVER commit a migration you have not tested from an empty database.**

Need to change something you already migrated? Write a **new** migration that
changes the schema forward (add a column, drop a table, backfill data, ...). For
example, to rename a column in SQLite: add the new column, copy the data, then
drop the old one — in new migration files.

### Why migrations are not idempotent

A migration is applied **exactly once**. The runner guarantees this by recording
each applied version in `schema_migrations` and by executing the migration and
its version record inside a single transaction: a migration that fails is not
recorded, and an applied migration is never run again.

Because of that, migrations must **not** be written to be idempotent. In
particular, do not use `CREATE TABLE IF NOT EXISTS`, `CREATE INDEX IF NOT
EXISTS`, or similar:

* If an object already exists with a **different** structure, `IF NOT EXISTS`
  silently succeeds and leaves the database in a state that does not match the
  migrations. You want that to fail loudly, not to be masked.
* Many operations cannot be made idempotent at all (adding a column, backfilling
  data, inserting seed rows), so idempotency is not a general solution and gives
  a false sense of safety.
* The migration history in `schema_migrations` — not the SQL — is the source of
  truth for what has been applied.

Write each migration as the single, unconditional change from the previous schema
version to the next.

### Resetting your development database

Since migrations only go forward, the easiest way to start over while developing
is to empty the database and restart the server. The full set of migrations is
re-applied to a fresh database.

* **SQLite**: delete the database file (by default `/tmp/decaf.db`, or whatever
  `--db-filename` points to).
* **PostgreSQL**: drop and recreate the database (for example `DROP DATABASE decaf;`
  and `CREATE DATABASE decaf;` from `psql`), or
  simply recreate the PostgreSQL container if you run it with Docker/Podman.

### Advanced: using an ORM or sqlc

> This is meant for advanced students only. In the WASA course you are expected to use
> the migration system described above. But if you are already an experienced backend developer and want to use advanced tools (at your own risk), keep reading

The migration system is the default and recommended data layer. If you already
know what you are doing and want a different one, the project is designed so you
can plug it in without touching the API handlers:

* All data access goes through the `database.AppDatabase` interface
  (`service/database/database.go`). Handlers depend on that interface, not on the
  concrete SQLite or PostgreSQL implementation.
* **sqlc**: point its `schema` setting at the `migrations/` directory of your database package (keep them
  forward-only, with no "down" files) and its `queries` setting at a directory of
  your own `.sql` files. Commit and vendor the generated code, then write a small
  adapter type that implements `AppDatabase` and pass it to `api.New` in
  `cmd/webapi/main.go`.
* **GORM**: keep running the SQL migrations first and **do not** enable
  `AutoMigrate`: the migrations are the only source of truth for the schema. Map
  your models to the migrated tables and implement `AppDatabase` as a thin
  wrapper.
* In both cases keep calling `database.Migrate` at startup, as `sqlite.New` and
  `postgres.New` do. The handlers and the `api` package should not need any change.

This section is intentionally just a pointer in the right direction: if you take
this route, you are expected to work out the details yourself.

However, keep in mind that any significant changes to the Fantastic Coffee project structure or template are made at your own risk. If such changes introduce problems or incompatibilities, students are responsible for resolving them, and we cannot provide support for issues caused by substantial modifications to the provided template.

## Object storage

Some applications need to store files (photos, attachments, ...). Files do not belong in the database: they go in an
*object storage*, a service that stores each file (an *object*) under a name (a *key*, for example
`photos/42.jpg`) inside a container (a *bucket*). The database stores only the key.

This project works with any service compatible with the Amazon S3 API, for example
[Garage](https://garagehq.deuxfleurs.fr/), MinIO or AWS S3, using the [minio-go](https://github.com/minio/minio-go)
library:

* `service/objectstorage/objectstorage.go` defines the `ObjectStorage` interface. The API handlers depend only on this
  interface.
* `service/objectstorage/s3` implements it for S3-compatible services.

Object storage is **optional**: it is enabled only when the S3 endpoint is configured. Otherwise the web API starts
without it, and the `Storage` field of `api.Config` is `nil`: handlers that use it must check for `nil` first (the
example replies `503 Service Unavailable`).

### Configuration

| Environment variable | Flag | Example | Description |
|---|---|---|---|
| `CFG_S3_ENDPOINT` | `--s3-endpoint` | `http://localhost:3900` | S3 API endpoint, used both by the backend and for generating presigned URLs. It must start with `http://` or `https://`. If empty, object storage is disabled |
| `CFG_S3_REGION` | `--s3-region` | `eu-south-1` | Region used for signing requests (for Garage, it is the `s3_region` value in `garage.toml`) |
| `CFG_S3_BUCKET` | `--s3-bucket` | `wasatextplus` | Name of the bucket where files are stored. It must already exist |
| `CFG_S3_ACCESS_KEY_ID` | `--s3-access-key-id` | `GK3515373e4c851ebaad366558` | Access key ID |
| `CFG_S3_SECRET_ACCESS_KEY` | `--s3-secret-access-key` | `7d37d093...` | Secret access key |

or, in the configuration file:

```yaml
s3:
  endpoint: http://localhost:3900
  region: eu-south-1
  bucket: wasatextplus
  accesskeyid: GK3515373e4c851ebaad366558
  secretaccesskey: 7d37d093...
```

When the endpoint is set, all the other values are required, and the web API checks at startup that it can reach the
bucket: if anything is wrong (a missing value, a wrong region or key, a bucket that does not exist), the server does
not start and the error explains why.

### Presigned URLs: how files are uploaded and downloaded

Files are not sent through the web API. Instead, the web API gives the client a *presigned URL*: a temporary URL,
signed with the access key, that allows the client (for example, the browser) to upload or download one specific file
directly to/from the object storage. The web API only decides who can do what, and the files never pass through it.

The example endpoints `PUT /example-photo` and `GET /example-photo` (`service/api/set-example-photo.go` and
`service/api/get-example-photo.go`) upload and download a single JPEG photo, stored under the fixed key
`example-photo`:

1. The client asks to upload a photo, sending its size:
   `PUT /example-photo` with body `{"size": 123456}`. The web API checks the size (maximum 10 MB) and replies with
   `{"url": "...", "contentType": "image/jpeg"}`.
2. The client uploads the photo directly to the object storage, with an HTTP `PUT` to that URL:

   ```js
   await fetch(reply.url, {method: "PUT", body: file, headers: {"Content-Type": reply.contentType}})
   ```
3. To show the photo, the client calls `GET /example-photo`, which replies with `{"url": "..."}` (or `404` if no photo
   has been uploaded yet), and uses the URL directly, for example as the `src` of an `<img>`.

Things to know:

* **The size and the content type are part of the upload URL signature**: the object storage rejects the upload
  (`403 Forbidden`) if the file has a different size or is sent with a different `Content-Type`. This is how the web API
  enforces the maximum size. The object storage does not check that the file is really a JPEG: the client declares it.
* **Presigned URLs expire** (after 5 minutes in the example): request a new one when needed, do not store them.
* **The endpoint must be reachable by whoever uses the URL.** Presigned URLs contain the address of
  `CFG_S3_ENDPOINT`, so it must be an address that the browser can reach (`http://localhost:3900` is fine while
  developing on your computer), not a name that works only inside a Docker network.
* In a real application, generate a new key for each file (for example, a random UUID) instead of a fixed one. After
  the client uploads the file, call `Stat` to check that the upload actually happened (and its size and type) before
  saving the key in the database.
* The interface also has `Put` and `Get`, which upload and download through the web API. Use them only for files that
  the backend creates itself (for example, generated thumbnails).

### Bucket CORS (required for browsers)

When the web UI uploads or downloads with `fetch`, the browser talks to the object storage, which is on a different
origin than the web UI. The browser allows it only if the bucket has a CORS rule. Without it, uploads fail in the
browser (`TypeError: Failed to fetch`, and in the console *"Response to preflight request doesn't pass access control
check"*) even though they work with `curl`.

Save this rule in `cors.json`:

```json
{
  "CORSRules": [
    {
      "AllowedOrigins": ["*"],
      "AllowedMethods": ["GET", "PUT"],
      "AllowedHeaders": ["content-type"],
      "MaxAgeSeconds": 3600
    }
  ]
}
```

and apply it with any S3 client, for example the [AWS CLI](https://aws.amazon.com/cli/) (it can run in a container,
using the `amazon/aws-cli` image):

```shell
AWS_ACCESS_KEY_ID=<key ID> AWS_SECRET_ACCESS_KEY=<secret key> AWS_DEFAULT_REGION=eu-south-1 \
    aws --endpoint-url http://localhost:3900 s3api put-bucket-cors --bucket wasatextplus --cors-configuration file://cors.json
```

**Write `content-type` in lowercase.** Garage compares header names case-sensitively, and browsers send them in
lowercase: with `Content-Type`, the browser upload fails. `"AllowedOrigins": ["*"]` allows any web page, like the CORS
policy of the web API (`cmd/webapi/cors.go`); you can restrict it to the address of your web UI.

### Running Garage on your computer

If you are given a ready Garage instance, you only need its endpoint, region, bucket and access key. To run your own
for development, create a `garage.toml` file:

```toml
metadata_dir = "/var/lib/garage/meta"
data_dir = "/var/lib/garage/data"
db_engine = "sqlite"
replication_factor = 1
rpc_bind_addr = "[::]:3901"
rpc_public_addr = "127.0.0.1:3901"
rpc_secret = "<the output of: openssl rand -hex 32>"

[s3_api]
s3_region = "eu-south-1"
api_bind_addr = "[::]:3900"
root_domain = ".s3.garage.localhost"
```

then start Garage and create a bucket and an access key:

```shell
docker run -d --name garage -p 3900:3900 -v ./garage.toml:/etc/garage.toml:ro dxflrs/garage:v2.1.0
docker exec garage /garage node id -q        # prints <node ID>@...: use the part before @ below
docker exec garage /garage layout assign -z dc1 -c 1G <node ID>
docker exec garage /garage layout apply --version 1
docker exec garage /garage bucket create wasatextplus
docker exec garage /garage key create wasatextplus-key   # prints the key ID and the secret key
docker exec garage /garage bucket allow --read --write --owner wasatextplus --key wasatextplus-key
```

Finally, apply the CORS rule (see above). The data is stored inside the container: it is lost when the container is
removed.

### Removing object storage

If your project does not need to store files, you can remove object storage completely:

1. Delete the `service/objectstorage/` directory.
2. Delete the example: `service/api/get-example-photo.go`, `service/api/set-example-photo.go`, and their two routes in
   `service/api/api-handler.go`.
3. In `service/api/api.go`, remove the `Storage` field from `Config`, the `storage` field from `_router` (and the line
   that sets it in `New`), and the `objectstorage` import.
4. In `cmd/webapi/main.go`, remove the block that starts the object storage, the `Storage: storage` line passed to
   `api.New`, and the `objectstorage` and `s3` imports.
5. In `cmd/webapi/load-configuration.go`, remove the `S3` block.
6. Run `go mod tidy` and then `go mod vendor` to remove `minio-go` from `go.mod` and `vendor/`, and check that
   `go build ./...` still works.

## Node/YARN vendoring

This repository uses `yarn` and a vendoring technique that exploits the ["Offline mirror"](https://yarnpkg.com/features/caching). As for the Go vendoring, the dependencies are inside the repository.

You should commit the files inside the `.yarn` directory.

## How to set up a new project from this template

You need to:

* Change the Go module path to your module path in `go.mod`, `go.sum`, and in `*.go` files around the project
* Rewrite the API documentation `doc/api.yaml`
* If no web frontend is expected, remove `webui` and `cmd/webapi/register-webui.go`
* Update top/package comment inside `cmd/webapi/main.go` to reflect the actual project usage, goal, and general info
* Pick the right database between SQLite, PostgreSQL, or support both. If you only need to support one, delete the other implementation (see "Choosing a database")
* If you need to store files, configure the object storage; otherwise, you can remove it (see "Object storage")
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
