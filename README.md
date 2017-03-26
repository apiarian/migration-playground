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


## Shiny API

The Shiny API, found in the [shiny-api](./shiny-api/) directory, is a new
stream-backed API. It handles `Thing` entities with list/get/create/update
actions. It also handles command entities with get actions. Entities are kept in
a kafka stream and cached locally as needed. The system is broken into two
parts, the API itself [shiny-api/api](./shiny-api/api/) and the Updater tool
[shiny-api/updater](./shiny-api/updater/). The API creates commands and watches
for responses on a kafka topic, while the updater processes the business logic
and deals with any data coordination that may be required with the Original API.

### Running the Shiny API

The API command can be launched by executing `go run shiny-api/api/*.go`. The
Updater command can be launched by executing `go run shiny-api/updater/*.go`.

See the respective `-help` output for command-line arguments.

**NOTE:** Kafka needs to be available for the API to function.

### Schema

`Thing` entities have the following schema:

```
{
	id: string (numeric)
	name: string
	foo: float
	created_on: string(timestamp)
	updated_on: string(timestamp)
	version: string (numeric)
}
```

The `name` and `foo` fields are required when creating `Things` with the `POST
api/things/` endpoint or updated with the `POST api/things/:id` endpoint. The
`version` field is also required when updating a `Thing`, and must match the
current value of the `version`. The other fields are read-only and will be
available at `GET api/things/` and `GET api/things/:id` for getting a single
`Thing` and listing all the `Things`, respectively.

**NOTE**: schema is similar and mappable (with slight loss in the `foo` field)
to the Original API. Even though the `id` and the `version` fields have been
turned into "opaque strings", they will need to be numeric for the duration of
the multi-master dance of the two APIs.

Error responses have the following schema:

```
{
	error-message: string
}
```

Error responses have non-200 HTTP status codes.

`application/json` in, `application/json` out.


## Running Everything

Run the Original API, Shiny API, and Updater all coordinated with the right
topic names using `./run-apis.sh`.

**NOTE:** the script is rather aggressive.
