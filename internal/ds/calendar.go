package ds

import (
	"time"

	"google.golang.org/api/calendar/v3"
)

type Cal struct {
	srv   *calendar.Service
	calId string
}

func (c *Cal) CreateAppointment() {

}

func (c *Cal) GetEvents(startDate time.Time, days int) {
	endDate := startDate.AddDate(0, 0, days)
	events, err := c.srv.Events.List(c.calId).ShowDeleted(false).
		SingleEvents(true).TimeMin(startDate.Format(time.RFC3339)).TimeMax(endDate.Format(time.RFC3339)).OrderBy("startTime").Do()
	startDate.Weekday()
}
