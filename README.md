# Go EventStore

[![Go](https://github.com/aneshas/eventstore/actions/workflows/go.yml/badge.svg?branch=master)](https://github.com/aneshas/eventstore/actions/workflows/go.yml)
[![Coverage Status](https://coveralls.io/repos/github/aneshas/eventstore/badge.svg)](https://coveralls.io/github/aneshas/eventstore)
[![Go Report Card](https://goreportcard.com/badge/github.com/aneshas/eventstore)](https://goreportcard.com/report/github.com/aneshas/eventstore)

Embeddable EventStore implementation written in Go using gorm as an underlying persistence mechanism meaning it will work
with `almost` (tested sqlite and postgres) whatever underlying database gorm will support (just use the respective gorm driver).

## Features

- Appending (saving) events to a particular stream
- Reading events from the stream
- Reading all events
- Subscribing (streaming) all events from the event store (real-time)
- Fault-tolerant projection system (Projector)

## Upcoming

Add offset handling and retry mechanism to the default Projector.

## Example

I provided a simple [example](example/) that showcases basic usage with sqlite.
