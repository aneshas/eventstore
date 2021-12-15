# Eventstore Example

This example shows a simplistic but typical event-sourcing use case scenario.

It contains a single "aggregate" (Account) that produces a single account opening event, an accompanying repository implementation along with a simple console output projection which uses a projector that is included in the eventstore package.

## How to run

Run both `cmd/api/main.go` and `cmd/projections/main.go` in any order from the same directory (so they use the same sqlite db). This will start a simple http api on `localhost:8080` and run projection binary which will subscribe to the event store and wait for incoming events in order to process them.

In order to simulate the account being open simply hit `http://localhost:8080/accounts/open` from your browser and monitor the terminal window from which you have started the projections binary.
