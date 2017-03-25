# Migration Playground

A place to run experiments on migrating from one API backend to another using
Kafka topics to choreograph the multi-master dance.


## Local Kafka

### Setup

1. Download the [Confluent Open Source Platform](https://www.confluent.io).
1. Extract the tar or zip file in the `~/development/kafka` directory
1. Make a symlink to the latest confluent directory at `~/development/kafka/latest`

### Running Kafka

1. Open three terminal windows at the `migration-playground` directory
1. Run `make zookeeper` in the first and wait for zookeeper load
1. Run `make kafka` in the second and wait for kafka to load
1. Run `make registry` in the third and wait for the schema registry to load
1. When you're ready to shut things down, send an Interrupt (`^C`) to each of
the terminal windows in reverse order


## Original API

The Original API, found in the [original-api](./original-api/) directory, is a
traditional permanent-storage-backed API. It handles `Thing` entities with
list/get/create/update actions. Entities are stored in a persistent storage
medium of some kind. The API also publishes the latest state of each entity to a
kafka topic.

### Running the Original API

The API can be launched by executing `go run original-api/*.go`.

See `go run original-api/*.go -help` for command-line arguments.

**NOTE:** Kafka needs to be available for the API to function.

### Schema

`Thing` entities have the following schema:

```
{
	id: integer
	name: string
	foo: integer
	created_on: string(timestamp)
	updated_on: string(timestamp)
	version: integer
}
```

The `name` and `foo` fields are required when creating `Things` with the `POST
api/things/` endpoint or updated with the `POST api/things/:id` endpoint. The
`version` field is also required when updating a `Thing`, and must match the
current value of the `version`. The other fields are read-only and will be
available at `GET api/things/` and `GET api/things/:id` for getting a single
`Thing` and listing all the `Things`, respectively.

Error responses have the following schema:

```
{
	error-message: string
}
```

Error responses have non-200 HTTP status codes.

`application/json` in, `application/json` out.
