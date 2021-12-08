package eventstore

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/mattn/go-sqlite3"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

const (
	InitialStreamVersion int = 0
)

var (
	ErrStreamNotFound         = errors.New("stream not found")
	ErrConcurrencyCheckFailed = errors.New("optimistic concurrency check failed: stream version exists")
)

type EventData struct {
	Event interface{}
	Meta  map[string]string
	Type  string
	// TODO Add CreatedAt(), Version(),  Offset?
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

type gormEvent struct {
	gorm.Model
	Type          string
	Stream        string `gorm:"index:idx_optimistic_check,unique;index"`
	StreamVersion int    `gorm:"index:idx_optimistic_check,unique"`
	Data          string
	Meta          string
}

type appendEventsConfig struct {
	meta map[string]string
}

type AppendEventsOpt func(appendEventsConfig) appendEventsConfig

func WithMetaData(meta map[string]string) AppendEventsOpt {
	return func(cfg appendEventsConfig) appendEventsConfig {
		cfg.meta = meta

		return cfg
	}
}

func (e *EventStore) AppendStream(
	ctx context.Context,
	stream string,
	expectedVer int,
	evts []interface{},
	opts ...AppendEventsOpt) error {

	cfg := appendEventsConfig{}

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

	outEvts := make([]EventData, len(evts))

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

		outEvts[i] = EventData{
			Event: data,
			Type:  evt.Type,
			Meta:  meta,
		}
	}

	return outEvts, nil
}
