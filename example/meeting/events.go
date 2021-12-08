package meeting

import "time"

type MeetingScheduled struct {
	MeetingID   string
	ScheduledOn time.Time
}

type MeetingPostponed struct {
	MeetingID   string
	PostponedBy time.Duration
}
