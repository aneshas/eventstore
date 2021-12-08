package main

import (
	"context"

	"github.com/aneshas/goddd"
	"github.com/aneshas/goddd/eventstore"
	"github.com/aneshas/goddd/example/meeting"
)

type MeetingRepo struct {
	store *eventstore.EventStore
}

func (r *MeetingRepo) Save(ctx context.Context, m *meeting.Meeting) error {
	var evts []interface{}

	for _, evt := range m.Events() {
		evts = append(evts, evt)
	}

	return r.store.AppendStream(
		context.Background(),
		m.ID().String(),
		m.Version(),
		evts,
		// eg. read meta from ctx
		eventstore.WithMetaData(nil),
	)
}

func (r *MeetingRepo) FindByID(id meeting.MeetingID) (*meeting.Meeting, error) {
	evts, err := r.store.ReadStream(context.Background(), id.String())
	if err != nil {
		return nil, err
	}

	var dddEvts []goddd.DomainEvent

	for _, evt := range evts {
		// eg. extract additional props (CreatedAt etc...)
		dddEvts = append(dddEvts, evt.Event)
	}

	return meeting.New(dddEvts)
}
