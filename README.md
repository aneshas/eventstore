# Go EventStore

[![Go](https://github.com/aneshas/eventstore/actions/workflows/go.yml/badge.svg?branch=master)](https://github.com/aneshas/eventstore/actions/workflows/go.yml)
[![Coverage Status](https://coveralls.io/repos/github/aneshas/eventstore/badge.svg)](https://coveralls.io/github/aneshas/eventstore)
[![Go Report Card](https://goreportcard.com/badge/github.com/aneshas/eventstore)](https://goreportcard.com/report/github.com/aneshas/eventstore)

Embeddable SQL EventStore + Aggregate Abstraction written in Go using gorm as an underlying persistence mechanism meaning - it will work
with `almost` (tested sqlite and postgres) whatever underlying database gorm will support (just use the respective gorm driver - sqlite and postgres provided).

It is also equiped with a fault-tolerant projection system that can be used to build read models for testing purposes and is also ready for 
production workloads in combination with [Ambar.cloud](https://ambar.cloud/) using the provided ambar package.

## Features

- Appending (saving) events to a particular stream
- Reading events from the stream
- Reading all events
- Subscribing (streaming) all events from the event store (real-time - polling)
- Aggregate root abstraction to manage rehydration and event application
- Generic aggregate store implementation used to read and save aggregates (events)
- Fault-tolerant projection system (Projector) which can be used to build read models for testing purposes
- [Ambar.cloud](https://ambar.cloud/) data destination (projection) integration for production projection workloads - see [example](example/)

## Example

I provided a simple [example](example/) that showcases basic usage with sqlite, or check out [cqrs clinic example](https://github.com/aneshas/cqrs-clinic)
