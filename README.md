# Go EventStore

[![Go](https://github.com/aneshas/eventstore/actions/workflows/go.yml/badge.svg?branch=master)](https://github.com/aneshas/eventstore/actions/workflows/go.yml)
[![Coverage Status](https://coveralls.io/repos/github/aneshas/eventstore/badge.svg)](https://coveralls.io/github/aneshas/eventstore)
[![Go Report Card](https://goreportcard.com/badge/github.com/aneshas/eventstore)](https://goreportcard.com/report/github.com/aneshas/eventstore)

# TODO 
- [ ] store tests
- [ ] add postgres tests with test-containers (with flag?)
- [ ] upgrade packages
- [x] autogenerate ID in aggregate
- [ ] WithDB option to pass db connection - eg. for encore?
- [ ] complete example with echo and mutation
- [ ] for projections - ignore missing json types - don't throw error (this implies projection is not interested)
- [ ] alternate On method with Event and alternate apply for id
- [x] correlation from context (helper methods in aggregate to set correlation and meta)
- [ ] projections only for simple local testing simplify and document
- [x] json encoding/decoding for events types - better way?

Embeddable EventStore implementation written in Go using gorm as an underlying persistence mechanism meaning it will work
with `almost` (tested sqlite and postgres) whatever underlying database gorm will support (just use the respective gorm driver).

## Features

- Appending (saving) events to a particular stream
- Reading events from the stream
- Reading all events
- Subscribing (streaming) all events from the event store (real-time)
- Fault-tolerant projection system (Projector)

## Example

I provided a simple [example](example/) that showcases basic usage with sqlite.
