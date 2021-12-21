# Go EventStore

[![Go](https://github.com/aneshas/eventstore/actions/workflows/go.yml/badge.svg?branch=master)](https://github.com/aneshas/eventstore/actions/workflows/go.yml)
[![Coverage Status](https://coveralls.io/repos/github/aneshas/eventstore/badge.svg)](https://coveralls.io/github/aneshas/eventstore)
[![Go Report Card](https://goreportcard.com/badge/github.com/aneshas/eventstore)](https://goreportcard.com/report/github.com/aneshas/eventstore)

Embeddable eventstore implementation written in Go using sqlite as an underlying persistence mechanism.

## Why sqlite?

In the future I might choose to abstract the persistence mechanism away and add more options such as bolt or smth, but for now, I decided to go with sqlite because I think it does the job well and especially now that we have [litestream](https://github.com/benbjohnson/litestream) thanks to the amazing @benbjohnson it becomes even more flexible.

## Features

- Appending (saving) events to a particular stream
- Reading events from the stream
- Reading all events
- Subscribing (streaming) all events from the event store (real-time)
- Fault-tolerant projection system (Projector)

## Upcoming

Add offset handling and retry mechanism to the default Projector.

## Example

I provided a simple [example](example/) that showcases basic usage.
