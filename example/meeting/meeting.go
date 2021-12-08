package meeting

import (
	"time"

	"github.com/aneshas/goddd"
)

func New(evts []goddd.DomainEvent) (*Meeting, error) {
	f := Meeting{}

	f.Init(&f, evts)

	return &f, nil
}

func Schedule(id MeetingID, date time.Time) (*Meeting, error) {
	f := Meeting{}

	f.ApplyEvent(&f, &MeetingScheduled{
		MeetingID:   id.String(),
		ScheduledOn: date,
	})

	return &f, nil
}

func NewMeetingID() MeetingID {
	return MeetingID{
		id: "new-meeting-123",
	}
}

type MeetingID struct {
	id string
}

func (id MeetingID) String() string {
	return id.id
}

type Meeting struct {
	goddd.BaseAggregateRoot

	id          MeetingID
	scheduledOn time.Time
}

func (m *Meeting) ID() goddd.DomainID { return m.id }

func (m *Meeting) PostponeBy(d time.Duration) {
	m.ApplyEvent(m, &MeetingPostponed{
		MeetingID:   m.id.String(),
		PostponedBy: d,
	})
}

func (m *Meeting) OnMeetingScheduled(e *MeetingScheduled) {
	m.id = MeetingID{
		id: e.MeetingID,
	}

	m.scheduledOn = e.ScheduledOn
}

func (m *Meeting) OnMeetingPostponed(e *MeetingPostponed) {
	m.scheduledOn = m.scheduledOn.Add(e.PostponedBy)
}
