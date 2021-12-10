// Package eventstore provides a simple light-weight event store implementation
// that uses sqlite as a backing storage.
// Apart from the event store, mechanisms for building projections and
// working with aggregate roots are provided
package eventstore

import (
	"context"
	"encoding/json"
	"errors"
	"io"

	"github.com/mattn/go-sqlite3"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var (
	// ErrStreamNotFound indicates that the requested stream does not exist in the event store
	ErrStreamNotFound = errors.New("stream not found")

	// ErrConcurrencyCheckFailed indicates that stream entry related to a particular version already exists
	ErrConcurrencyCheckFailed = errors.New("optimistic concurrency check failed: stream version exists")

	// ErrSubscriptionClosedByClient is produced by sub.Err if client cancels the subscription using sub.Close()
	ErrSubscriptionClosedByClient = errors.New("subscription closed by client")
)

// EncodedEvt represents encoded event used by a specific encoder implementation
type EncodedEvt struct {
	Data string
	Type string
}

// Encoder is used by the event store in order to correctly marshal
// and unmarshal event types
type Encoder interface {
	Encode(interface{}) (*EncodedEvt, error)
	Decode(*EncodedEvt) (interface{}, error)
}

// EventData holds stored event data and meta data
type EventData struct {
	Event interface{}
	Meta  map[string]string
	Type  string
	// TODO Add CreatedAt(), Version(),  Offset, Stream
}

// New construct new event store
// dbname - a path to sqlite database on disk
// enc - a specific encoder implementation (see bundled JsonEncoder)
func New(dbname string, enc Encoder) (*EventStore, error) {
	db, err := gorm.Open(sqlite.Open(dbname), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	db.AutoMigrate(&gormEvent{})

	return &EventStore{
		db:  db,
		enc: enc,
	}, nil
}

// EventStore represents a sqlite event store implementation
type EventStore struct {
	db  *gorm.DB
	enc Encoder
}

// Close should be called as a part of cleanup process
// in order to close the underlying sql connection
func (es *EventStore) Close() error {
	sqlDB, err := es.db.DB()
	if err != nil {
		return err
	}

	return sqlDB.Close()
}

type gormEvent struct {
	gorm.Model
	Type          string
	Stream        string `gorm:"index:idx_optimistic_check,unique;index"`
	StreamVersion int    `gorm:"index:idx_optimistic_check,unique"`
	Data          string
	Meta          string
}

type appendStreamConfig struct {
	meta map[string]string
}

type appendStreamOpt func(appendStreamConfig) appendStreamConfig

// WithMetaData is an AppendStream option that can be used to
// associate arbitrary meta data to a batch of events to store
func WithMetaData(meta map[string]string) appendStreamOpt {
	return func(cfg appendStreamConfig) appendStreamConfig {
		cfg.meta = meta

		return cfg
	}
}

const (
	// InitialStreamVersion can be used as an initial expectedVer for
	// new streams (as an argument to AppendStream)
	InitialStreamVersion int = 0
)

// AppendStream will encode provided event slice and try to append them to
// an indicated stream. If the stream does not exist it will be created.
// If the stream already exists an optimistic concurrency check will be performed
// using a compound key (stream-expectedVer).
// expectedVer should be InitialStreamVersion for new streams and the latest
// stream version for existing streams, otherwise a concurrency error
// will be raised
func (e *EventStore) AppendStream(
	ctx context.Context,
	stream string,
	expectedVer int,
	evts []interface{},
	opts ...appendStreamOpt) error {

	cfg := appendStreamConfig{}

	for _, opt := range opts {
		cfg = opt(cfg)
	}

	events := make([]gormEvent, len(evts))

	m, err := json.Marshal(cfg.meta)
	if err != nil {
		return err
	}

	for i, evt := range evts {
		encoded, err := e.enc.Encode(evt)
		if err != nil {
			return err
		}

		expectedVer++

		events[i] = gormEvent{
			Stream:        stream,
			StreamVersion: expectedVer,
			Data:          encoded.Data,
			Meta:          string(m),
			Type:          encoded.Type,
		}
	}

	tx := e.db.Create(&events)

	if e, ok := tx.Error.(sqlite3.Error); ok && e.Code == 19 {
		return ErrConcurrencyCheckFailed
	}

	return tx.Error
}

type readAllConfig struct {
	offset    int
	batchSize int
}

type readAllOpt func(readAllConfig) readAllConfig

// WithOffset is a ReadAll option that indicates an offset in the event store
// from wich to start reading events (exclusive)
func WithOffset(offset int) readAllOpt {
	return func(cfg readAllConfig) readAllConfig {
		cfg.offset = offset

		return cfg
	}
}

// WithBatchSize is a ReadAll option that specifies the read batch size (limit)
// when reading events from the event store
func WithBatchSize(size int) readAllOpt {
	return func(cfg readAllConfig) readAllConfig {
		cfg.batchSize = size

		return cfg
	}
}

// Subscription represents ReadAll subscription that is used for streaming
// incoming events
type Subscription struct {
	// Err chan will produce any errors that might occurr while reading events
	// If Err produces io.EOF error, that indicates that we have caught up
	// with the event store and that there are no more events to read after which
	// the subscription itself will continue polling the event store for new events
	// each time we empty the Err channel. This means that reading from Err (in
	// case of io.EOF) can be strategically used in order to achieve backpressure
	Err       chan error
	EventData chan EventData

	close chan struct{}
}

func (s Subscription) Close() {
	s.close <- struct{}{}
}

// ReadAll will create a subscription which will stream all events in an
// orderly fashion. This mechanism should probably be mostly useful for
// building projections
func (e *EventStore) ReadAll(ctx context.Context, opts ...readAllOpt) (Subscription, error) {
	cfg := readAllConfig{
		offset:    0,
		batchSize: 100,
	}

	// TODO parse opts

	sub := Subscription{
		Err:       make(chan error, 1),
		EventData: make(chan EventData, cfg.batchSize),
		close:     make(chan struct{}, 1),
	}

	go func() {
		var done error

		for {
			select {
			case <-sub.close:
				sub.Err <- ErrSubscriptionClosedByClient

				return
			case <-ctx.Done():
				sub.Err <- ctx.Err()

				return
			default:
				// Make sure client reads all buffered events
				if done != nil {
					if len(sub.EventData) != 0 {
						break
					}

					sub.Err <- done

					return
				}

				var evts []gormEvent

				if err := e.db.
					Where("id > ?", cfg.offset).
					Order("id asc").
					Limit(cfg.batchSize).
					Find(&evts).Error; err != nil {
					done = err

					break
				}

				if len(evts) == 0 {
					sub.Err <- io.EOF

					break
				}

				cfg.offset = cfg.offset + len(evts)

				decoded, err := e.decodeEvts(evts)
				if err != nil {
					done = err

					break
				}

				for _, evt := range decoded {
					sub.EventData <- evt
				}
			}
		}
	}()

	return sub, nil
}

// ReadStream will read all events associated with provided stream
// If there are no events stored for a given stream ErrStreamNotFound will be returned
func (e *EventStore) ReadStream(ctx context.Context, stream string) ([]EventData, error) {
	var evts []gormEvent

	if err := e.db.
		Where("stream = ?", stream).
		Order("id asc").
		Find(&evts).Error; err != nil {

		return nil, err
	}

	if len(evts) == 0 {
		return nil, ErrStreamNotFound
	}

	return e.decodeEvts(evts)
}

func (e *EventStore) decodeEvts(evts []gormEvent) ([]EventData, error) {
	out := make([]EventData, len(evts))

	for i, evt := range evts {
		data, err := e.enc.Decode(&EncodedEvt{
			Data: evt.Data,
			Type: evt.Type,
		})
		if err != nil {
			return nil, err
		}

		var meta map[string]string

		err = json.Unmarshal([]byte(evt.Meta), &meta)
		if err != nil {
			return nil, err
		}

		out[i] = EventData{
			Event: data,
			Type:  evt.Type,
			Meta:  meta,
		}
	}

	return out, nil
}
