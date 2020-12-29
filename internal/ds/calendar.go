package ds

import (
	"fmt"
	"time"

	"google.golang.org/api/calendar/v3"
)

type Cal struct {
	srv   *calendar.Service
	calId string
}

func NewCal(srv *calendar.Service, calId string) *Cal {
	return &Cal{srv: srv, calId: calId}
}

func (c *Cal) CreateAppointment(startTime time.Time, mins int) error {
	endTime := startTime.Add(time.Minute * time.Duration(mins))
	event := &calendar.Event{
		Description: "consultatie",
		Summary:     "consultatie",
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

func (c *Cal) GetEvents(startDate time.Time, days int) {
	//endDate := startDate.AddDate(0, 0, days)
	//events, err := c.srv.Events.List(c.calId).ShowDeleted(false).
	//	SingleEvents(true).TimeMin(startDate.Format(time.RFC3339)).TimeMax(endDate.Format(time.RFC3339)).OrderBy("startTime").Do()
}
