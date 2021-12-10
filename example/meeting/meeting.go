package meeting

import (
	"time"

	"github.com/aneshas/eventstore"
)

func New(id MeetingID, date time.Time) (*Meeting, error) {
	m := Meeting{}

	m.Init(&m)

	err := m.Apply(
		MeetingScheduled{
			MeetingID:   id.String(),
			ScheduledOn: date,
		},
	)

	return &m, err
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
	eventstore.AggregateRoot

	id          MeetingID
	scheduledOn time.Time
}

func (m *Meeting) ID() MeetingID { return m.id }

func (m *Meeting) PostponeBy(d time.Duration) error {
	return m.Apply(
		MeetingPostponed{
			MeetingID:   m.id.String(),
			PostponedBy: d,
		},
	)
}

func (m *Meeting) OnMeetingScheduled(e MeetingScheduled) {
	m.id = MeetingID{
		id: e.MeetingID,
	}

	m.scheduledOn = e.ScheduledOn
}

func (m *Meeting) OnMeetingPostponed(e MeetingPostponed) {
	m.scheduledOn = m.scheduledOn.Add(e.PostponedBy)
}
