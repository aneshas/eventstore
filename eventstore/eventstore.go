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

const (
	InitialStreamVersion int = 0
)

var (
	ErrStreamNotFound             = errors.New("stream not found")
	ErrConcurrencyCheckFailed     = errors.New("optimistic concurrency check failed: stream version exists")
	ErrSubscriptionClosedByClient = errors.New("subscription closed by client")
)

type EventData struct {
	Event interface{}
	Meta  map[string]string
	Type  string
	// TODO Add CreatedAt(), Version(),  Offset, Stream
}

type EncodedEvt struct {
	Data string
	Type string
}

type Encoder interface {
	Encode(interface{}) (*EncodedEvt, error)
	Decode(*EncodedEvt) (interface{}, error)
}

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

type EventStore struct {
	db  *gorm.DB
	enc Encoder
}

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

func WithMetaData(meta map[string]string) appendStreamOpt {
	return func(cfg appendStreamConfig) appendStreamConfig {
		cfg.meta = meta

		return cfg
	}
}

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

func WithOffset(offset int) readAllOpt {
	return func(cfg readAllConfig) readAllConfig {
		cfg.offset = offset

		return cfg
	}
}

func WithBatchSize(size int) readAllOpt {
	return func(cfg readAllConfig) readAllConfig {
		cfg.batchSize = size

		return cfg
	}
}

type Subscription struct {
	// Err can also be used as backpressure
	// eg. every time io.EOF is read increase backoff
	Err       chan error
	EventData chan EventData

	close chan struct{}
}

func (s Subscription) Close() {
	s.close <- struct{}{}
}

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
