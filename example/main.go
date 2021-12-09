package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aneshas/goddd/eventstore"
	"github.com/aneshas/goddd/example/meeting"
)

var repo *MeetingRepo

func init() {
	store, err := eventstore.New(
		"meetingsdb",
		eventstore.NewJsonEncoder(
			meeting.MeetingScheduled{},
			meeting.MeetingPostponed{},
		),
	)
	checkErr(err)

	repo = &MeetingRepo{
		store: store,
	}
}

func main() {
	meetID := meeting.NewMeetingID()

	err := scheduleMeeting(meetID, time.Now())
	checkErr(err)

	err = postponeMeetingBy(meetID, time.Hour*11)
	checkErr(err)

	meet, err := repo.FindByID(meetID)
	checkErr(err)

	fmt.Printf("meet: %#v", meet)
}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func scheduleMeeting(id meeting.MeetingID, date time.Time) error {
	meet, err := meeting.Schedule(id, time.Now())
	if err != nil {
		return err
	}

	return repo.Save(context.Background(), meet)
}

func postponeMeetingBy(id meeting.MeetingID, d time.Duration) error {
	meet, err := repo.FindByID(id)
	if err != nil {
		return err
	}

	meet.PostponeBy(time.Hour * 10)

	return repo.Save(context.Background(), meet)
}

func runProjections(ctx context.Context, store *eventstore.EventStore) {
	p := eventstore.NewProjector(store)

	p.Add(
		NewMeetingsProjection(),
	)

	log.Fatal(p.Run(ctx))
}

// This would be barebones, we can also provide a default
// reflection based projector (later)
func NewMeetingsProjection( /* deps */ ) eventstore.Projection {
	return func(data eventstore.EventData) error {
		// switch data.Event.(type) {
		// case meeting.MeetingScheduled:
		// 	// Save new entry to table
		// 	break

		// case meeting.MeetingPostponed:
		// 	// Update existing entry
		// 	break
		// }

		return nil
	}
}
