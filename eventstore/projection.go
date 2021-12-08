package eventstore

type EventStreamer interface {
}

// TODO Add common middlewares and ability to wrap
// eg. retry mechanism, error handler...
type Projection func(EventData) error

func NewProjector(streamer EventStreamer, projections ...Projection) *Projector {
	return nil
}

type Projector struct {
}

// This would be barebones, we can also provide a default
// reflection based projector (later)
func NewMeetingsProjector( /* deps */ ) Projection {
	return func(data EventData) error {
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
