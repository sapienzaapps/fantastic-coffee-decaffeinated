# Development environment: PostgreSQL and Garage on your computer

The `demo/` directory contains everything needed to run, on your computer, the two external services that the web API
can use:

* **PostgreSQL**, if you use it instead of SQLite (see "Choosing a database" in `README.md`);
* **Garage**, an S3-compatible object storage, if your project stores files (see "Object storage" in `README.md`).

| File | What it is |
|---|---|
| `demo/docker-compose.yml` | Describes the containers: PostgreSQL and Garage |
| `demo/garage.toml` | The Garage configuration |
| `demo/config.yml` | The configuration file of the web API |

The web API itself is not started by Docker Compose: you keep running it with `go run ./cmd/webapi/`, and point it to
these services with its configuration file, `demo/config.yml` (step 7).

> **Development only.** The passwords and secrets in `demo/` are public: they are in this repository. Never use
> them, or this setup, for anything reachable from the Internet.

You need Docker with Docker Compose (Docker Desktop includes both). With Podman, write `podman compose` instead of
`docker compose`. **All the `docker compose` commands below must be run inside the `demo/` directory**, where
`docker-compose.yml` is.

## 1. Start the services

```shell
cd demo
docker compose up -d
```

`up` creates and starts the containers described in `docker-compose.yml`; `-d` runs them in the background. The first
time, it also downloads the images. Check that they are running:

```shell
docker compose ps
```

* PostgreSQL is now ready at `localhost:5432`, with user `decaf`, password `decaf` and database `decaf`: the official
  image creates them at the first start. If you need only PostgreSQL, skip to step 7.
* Garage is running at `http://localhost:3900`, but it cannot store anything yet: steps 2 to 6 prepare it. You do them
  **only once**: Garage keeps its configuration in its Docker volumes (see "Stopping and resetting").

## 2. The Garage command line tool

Garage is managed with the `garage` command, which is inside its container. `docker compose exec garage <command>`
runs a command in the running `garage` container, so every Garage command looks like this:

```shell
docker compose exec garage /garage status
```

The `status` command shows the Garage servers (called nodes) that make up your storage system.
Garage can run on multiple servers working together as one storage system, called a *cluster*, so your files are
stored on more than one server.
This provides redundancy: if one server fails, another still has a copy of your files.
In this case, we will run a cluster with only one node, inside a single container on your computer,
storing just one copy of each file (replication_factor = 1 in demo/garage.toml).
As a result, `/garage status` shows only one node:

```
==== HEALTHY NODES ====
ID                Hostname      Address         Tags  Zone  Capacity          DataAvail  ResyncQueue  Version
f2b6e5a027c3d935  4e06ed4a8f96  127.0.0.1:3901              NO ROLE ASSIGNED             0            v2.5.0
```

Note the node `ID` in the first column (yours is different): you need it in the next step.

## 3. Assign a role to the node (the *layout*)

Garage is designed to spread the data over many nodes, possibly in different places (*zones*). The *layout* says which
nodes store data, in which zone, and how much space each one offers. A new node has no role ("NO ROLE ASSIGNED"), and
until a layout exists Garage refuses every read and write.

Give the node a role in a zone named `dc1`, with 1 GB of capacity (replace `<node ID>` with the ID from step 2):

```shell
docker compose exec garage /garage layout assign -z dc1 -c 1G <node ID>
```

Layout changes are *staged* first, so that several of them can be reviewed together (`garage layout show`), and then
applied all at once. Apply this one:

```shell
docker compose exec garage /garage layout apply --version 1
```

`--version 1` is the number of the new layout: it starts from 1 and increases at every change (it protects you from
applying a layout different from the one you reviewed). Running `garage status` again now shows the node in zone `dc1`.

## 4. Create the bucket

A *bucket* is a container of objects (files); the web API stores all its files in one bucket. Create it:

```shell
docker compose exec garage /garage bucket create decaf
```

The name is up to you; this guide uses `decaf`, which is the value of `bucket` in the configuration file (step 7).

## 5. Create an access key and give it access to the bucket

Every S3 request must be signed with an *access key*, made of a public *key ID* and a *secret key*, like a username and
a password. The web API uses one to sign its requests and the presigned URLs. (Do not confuse it with the *key* of an
object, which is just its name in the bucket, like `example-photo`.) Create one:

```shell
docker compose exec garage /garage key create decaf-key
```

`decaf-key` is just a name for you. The output shows the two values that the web API needs:

```
Key ID:              GKc203fabc835ca2dc2af0b42a
Key name:            decaf-key
Secret key:          28cf42a5b3e7...
```

Copy the `Key ID` and the `Secret key`: they are `accesskeyid` and `secretaccesskey` in the configuration file
(step 7). If you lose them, `docker compose exec garage /garage key info decaf-key --show-secret` shows them again.

A new key cannot access any bucket. Allow it to read and write the bucket:

```shell
docker compose exec garage /garage bucket allow --read --write decaf --key decaf-key
```

## 6. Allow browsers to use the bucket (CORS)

With presigned URLs, the browser uploads and downloads files directly from Garage (`http://localhost:3900`), which is a
different *origin* (address) from the web UI. Browsers block such requests unless the server allows them with a *CORS
rule*. This is the rule for the bucket:

```json
{
  "AllowedOrigin": ["*"],
  "AllowedMethod": ["GET", "PUT"],
  "AllowedHeader": ["content-type"],
  "MaxAgeSeconds": 3600
}
```

It allows web pages from any origin (`*`) to download (`GET`) and upload (`PUT`), sending the `Content-Type` header,
and lets the browser remember the answer for an hour. `content-type` must be lowercase: Garage compares header names
case-sensitively, and browsers send them in lowercase.

The rule is set with Garage's *admin API*: an HTTP API to manage Garage, separate from the S3 API, available at
`http://localhost:3903`. Every request must carry the `admin_token` of `demo/garage.toml` in the `Authorization`
header, so you can send them with `curl`.

The admin API identifies a bucket by its ID, not by its name. Get the ID of the bucket (the long hexadecimal value on
the first line, after `Bucket:`):

```shell
docker compose exec garage /garage bucket info decaf
```

Then set the rule, replacing `<bucket ID>`:

```shell
curl -X POST -H 'Authorization: Bearer decaf-admin-token' -H 'Content-Type: application/json' 'http://localhost:3903/v2/UpdateBucket?id=<bucket ID>' -d '{"corsRules": [{"AllowedOrigin": ["*"], "AllowedMethod": ["GET", "PUT"], "AllowedHeader": ["content-type"], "MaxAgeSeconds": 3600}]}'
```

Garage replies with the description of the bucket, which now includes the rule in `corsRules`. To check the rule
later, ask for the same description:

```shell
curl -H 'Authorization: Bearer decaf-admin-token' 'http://localhost:3903/v2/GetBucketInfo?globalAlias=decaf'
```

Garage is now ready.

## 7. Connect the web API

The web API reads its settings from the configuration file `demo/config.yml`. In that file, every option is commented
out (with `#`), so the defaults are used. Uncomment the blocks of the services you want to use, and set their values:

* **PostgreSQL**: uncomment `db:`, `driver:` and `dsn:`. The values are already right for the PostgreSQL of
  `docker-compose.yml`.
* **Garage**: uncomment `s3:` and all the lines below it. `endpoint`, `region` and `bucket` are already right; replace
  `<key ID>` and `<secret key>` with the values from step 5.

For example, to use both:

```yaml
db:
  driver: postgres
  dsn: postgres://decaf:decaf@localhost:5432/decaf?sslmode=disable

s3:
  endpoint: http://localhost:3900
  region: garage
  bucket: decaf
  accesskeyid: GKc203fabc835ca2dc2af0b42a
  secretaccesskey: 28cf42a5b3e7...
```

* `region` must match `s3_region` in `demo/garage.toml`. In AWS, the region is the place where your data is stored
  (for example `eu-south-1` is Milan). Garage has no regions, but every S3 request is signed for one, so Garage
  accepts only the requests signed for the name in its configuration: here, `garage`.
* The secret key ends up in a file of the repository: this is acceptable only because it is a key of your local Garage,
  for development. Never write the secrets of a real service in a committed file.

Then start the web API from the repository root (not from `demo/`), telling it where the configuration file is with
`--config-path`:

```shell
go run ./cmd/webapi/ --config-path demo/config.yml
```

The log shows `using PostgreSQL database` and `using S3 object storage at http://localhost:3900, bucket decaf`. If
something is wrong, the web API stops at startup and the error says why.

Values in the configuration file take precedence over environment variables and command line flags: if you also set,
for example, `CFG_S3_BUCKET`, the value in the file wins.

To try the object storage, upload a photo with the example endpoints (see "Presigned URLs" in `README.md`), for
example from another terminal:

```shell
# Ask the web API for an upload URL, for a 1000 bytes photo
curl -X PUT -d '{"size": 1000}' http://localhost:3000/example-photo
# Upload 1000 bytes to that URL (copy it from the reply, between quotes, and replace each \u0026 with &)
head -c 1000 /dev/urandom | curl -X PUT -H 'Content-Type: image/jpeg' --data-binary @- '<upload URL>'
# Ask for a download URL: the photo is there
curl http://localhost:3000/example-photo
```

## Stopping and resetting

```shell
docker compose stop      # stop the containers; start them again with `docker compose up -d`
docker compose down      # stop and remove the containers; the data is kept
docker compose down -v   # stop and remove the containers, and DELETE all the data
```

The data of both services (the PostgreSQL tables, the Garage layout, bucket, key and files) lives in Docker *volumes*
(`postgres-data`, `garage-meta`, `garage-data` in `docker-compose.yml`), which survive `down`. After `down -v`, the
services start again empty: PostgreSQL recreates the user and the database by itself, while Garage must be prepared
again from step 2, and its new access key has new values.

## Troubleshooting

* **`port is already allocated` (or `address already in use`)**: another program uses port 5432 or 3900, for
  example a PostgreSQL installed on your computer. Change the first number of the port in `docker-compose.yml` (for
  example `"5433:5432"`) and use it in `dsn` or `endpoint` in `demo/config.yml`.
* **`checking the S3 bucket ... 400 Bad Request`**: the `region` in `demo/config.yml` does not match `s3_region` in
  `demo/garage.toml`.
* **`checking the S3 bucket ... Access Denied` or `Forbidden`**: wrong key ID or secret key, or the key is not allowed
  on the bucket (step 5).
* **`the S3 bucket "..." does not exist`**: the bucket name differs from `bucket` in `demo/config.yml`, or Garage was
  reset with `down -v` (redo steps 2 to 6).
* **Uploads work with `curl` but fail in the browser** (`TypeError: Failed to fetch`, and *"Response to preflight
  request doesn't pass access control check"* in the console): the CORS rule is missing or wrong (step 6).
* **`curl` cannot connect to `localhost:3903`**: Garage is not running; start it with `docker compose up -d`.
* **`Forbidden: Invalid bearer token` from `curl`**: the token after `Bearer` is not the `admin_token` of
  `demo/garage.toml`.
