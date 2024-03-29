package ds

import (
	"fmt"
	"time"

	"google.golang.org/api/calendar/v3"
)

const EventDescription = "med busy"

type Cal struct {
	srv   *calendar.Service
	calId string
}

func NewCal(srv *calendar.Service, calId string) *Cal {
	return &Cal{srv: srv, calId: calId}
}

func (c *Cal) CreateEvent(startTime time.Time, mins int) error {
	endTime := startTime.Add(time.Minute * time.Duration(mins))
	event := &calendar.Event{
		Description: EventDescription,
		Summary:     EventDescription,
		Start: &calendar.EventDateTime{
			DateTime: startTime.Format(time.RFC3339),
		},

		End: &calendar.EventDateTime{
			DateTime: endTime.Format(time.RFC3339),
		},
		Organizer: &calendar.EventOrganizer{
			DisplayName: "Auto booking",
			Id:          "auto-booking",
		},
	}

	_, err := c.srv.Events.Insert(c.calId, event).Do()
	if err != nil {
		return fmt.Errorf("could not create event %v, err: %w", event, err)
	}

	return nil
}

func (c *Cal) GetEvents(startDate time.Time, days int) (*calendar.Events, error) {
	endDate := startDate.AddDate(0, 0, days)
	return c.srv.Events.List(c.calId).ShowDeleted(false).
		SingleEvents(true).TimeMin(startDate.Format(time.RFC3339)).TimeMax(endDate.Format(time.RFC3339)).OrderBy("startTime").Do()
}

// CalendarEventsSlots contains as keys the unix times for events and as values the event ids.
type CalendarEventsSlots map[int64]string

func (c *Cal) GetEventSlots(startDate time.Time, days int) (map[int64]string, error) {
	ev, err := c.GetEvents(startDate, days)
	if err != nil {
		return nil, fmt.Errorf("failed gertting calendar events: %w", err)
	}
	res := make(map[int64]string)
	for _, calEv := range ev.Items {
		if calEv.Description != EventDescription {
			continue
		}
		evStart, err := time.Parse(time.RFC3339, calEv.Start.DateTime)
		if err != nil {
			return nil, fmt.Errorf("failed parsing calendar event start time: %w", err)
		}
		res[evStart.Unix()] = calEv.Id
	}
	return res, nil
}

func (c *Cal) DeleteEvent(id string) error {
	err := c.srv.Events.Delete(c.calId, id).Do()
	if err != nil {
		return fmt.Errorf("failed deleting existing event with id %s, error: %w", id, err)
	}
	return nil
}
