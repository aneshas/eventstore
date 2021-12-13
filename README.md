# Go EventStore

[![Go](https://github.com/aneshas/eventstore/actions/workflows/go.yml/badge.svg?branch=master)](https://github.com/aneshas/eventstore/actions/workflows/go.yml)
[![Coverage Status](https://coveralls.io/repos/github/aneshas/eventstore/badge.svg)](https://coveralls.io/github/aneshas/eventstore)
[![Go Report Card](https://goreportcard.com/badge/github.com/aneshas/eventstore)](https://goreportcard.com/report/github.com/aneshas/eventstore)

Embeddable eventstore implementation written in Go using sqlite as an underlying persistence mechanism.

## Features

- Appending (saving) events to a particular stream
- Reading events from the stream
- Reading all events
- Subscribing (streaming) all events from the event store (real-time)

## Upcoming

Configurable fault-tolerant projection system with offset handling.

## Example

I provided a simple [example](example/README.md) that showcases basic usage.
