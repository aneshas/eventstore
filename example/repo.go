package main

import (
	"context"

	"github.com/aneshas/eventstore"
	"github.com/aneshas/eventstore/example/meeting"
)

type MeetingRepo struct {
	store *eventstore.EventStore
}

func (r *MeetingRepo) Save(ctx context.Context, m *meeting.Meeting) error {
	return r.store.AppendStream(
		context.Background(),
		m.ID().String(),
		m.Version(),
		m.Events(),
		// eg. read meta from ctx
		eventstore.WithMetaData(nil),
	)
}

func (r *MeetingRepo) FindByID(id meeting.MeetingID) (*meeting.Meeting, error) {
	data, err := r.store.ReadStream(context.Background(), id.String())
	if err != nil {
		return nil, err
	}

	var evts []interface{}

	for _, evt := range data {
		// eg. extract additional props (CreatedAt etc...)
		evts = append(evts, evt.Event)
	}

	var m meeting.Meeting

	m.Init(&m, evts...)

	return &m, nil
}
