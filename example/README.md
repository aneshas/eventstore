# EventStore Example

This example shows a simplistic but typical event-sourcing use case scenario.

It contains a single "aggregate" (Account) that produces a set of events, with three projections - two of which make use of a built-in projector and flush after projection,
and one using built-in Ambar wrapper projection library

## How to run

Run both `cmd/api/main.go` and `cmd/projections/main.go` in any order from the same directory (so they use the same sqlite db). This will start a simple http api on `localhost:8080` and run projection binary which will subscribe to the event store and wait for incoming events in order to process them.

The api will run on localhost:8080 and following endpoints are available:
- `/accounts/open`
- `/accounts/:id/deposit/:amount`
- `/accounts/:id/withdraw/:amount`

Monitor the output of the projections binary in order to see the effects of console projection. In addition to that, the second projection should create a json file on disk containing created accounts (named `accounts.json`).

## Ambar
If you want to try it with Ambar - you will need to set up your db, host the `api` and `ambar_projections` binaries
somewhere (or run them locally and tunnel via ngrok for example) and set up Ambar resources yourself and run the api binary with `-pg` option. (will provide a local docker-compose configuration)

### Configuring your postgres db for Ambar:

```sql
CREATE USER replication REPLICATION LOGIN PASSWORD 'repl-pass';

GRANT CONNECT ON DATABASE "your-db-name" TO replication;

GRANT SELECT ON TABLE event TO replication;

CREATE PUBLICATION event_publication FOR TABLE event;
```

Ambar data-source configuration example for the above configuration:
```json
{
  "dataSourceType": "postgres",
  "description": "Eventstore",
  "dataSourceConfig": {
    "hostname": "your-db-host",
    "hostPort": "5432",
    "databaseName": "your-db-name",
    "tableName": "event",
    "publicationName": "event_publication",
    "columns": "id,sequence,type,data,meta,causation_event_id,correlation_event_id,stream_id,stream_version,occurred_on",
    "username": "replication",
    "password": "repl-pass",
    "serialColumn": "sequence",
    "partitioningColumn": "stream_id"
  }
}
```

After everything has been wired up you can run/deploy `ambar_projections` binary which exposes a single projection endpoint
(`/projections/accounts/v1`) to be used with Ambar data destination.

Set your data-destination username and password to `user` and `pass` respectively
