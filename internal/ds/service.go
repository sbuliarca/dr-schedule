package ds

import (
	"fmt"
	"log"
	"time"

	"github.com/sirupsen/logrus"
)

var LocalLoc *time.Location

func init() {
	var err error
	LocalLoc, err = time.LoadLocation("Europe/Bucharest")
	if err != nil {
		log.Panic(err)
	}
}

type Service struct {
	cal       *Cal
	webClient BusySlotsFetcher
}

type BusySlotsFetcher interface {
	GetBusySlots() (Slots, error)
}

func NewService(
	cal *Cal,
	webClient BusySlotsFetcher,
) *Service {
	return &Service{
		cal:       cal,
		webClient: webClient,
	}
}

type DaySchedule struct {
	StartHour    int
	StartMin     int
	EndHour      int
	EndMin       int
	SlotDuration int
}

// parameterize the days to schedule
const daysToSchedule = 31

const defaultSlotDuration = 30

type Slots map[int64]struct{}

func (s *Service) SyncSlots(startTime time.Time) error {
	logrus.Infof("Started syncing slots with start time %s", startTime.Format(time.RFC3339))
	startDate := time.Date(startTime.Year(), startTime.Month(), startTime.Day(), 0, 0, 0, 0, LocalLoc)

	busySlots, err := s.webClient.GetBusySlots()
	logrus.Debugf("found busy slots %v", busySlots)
	if err != nil {
		return err
	}

	calEvents, err := s.cal.GetEventSlots(startDate, daysToSchedule)
	if err != nil {
		return fmt.Errorf("failed getting the current calendar events, err: %w", err)
	}

	if err := s.syncToCalendar(busySlots, calEvents); err != nil {
		return err
	}

	logrus.Info("Finished syncing slots")
	return nil
}

func (s *Service) syncToCalendar(busySlots Slots, calEvents map[int64]string) error {
	/*	create events in calendar for the busy slots that don't have yet an event */
	for busySlot := range busySlots {
		_, hasEvent := calEvents[busySlot]
		if !hasEvent {
			if err := s.cal.CreateEvent(time.Unix(busySlot, 0), defaultSlotDuration); err != nil {
				return err
			}
		}
	}

	/*	remove calendar events that are not still in the busy slots */
	for calSlot, eventID := range calEvents {
		_, isStillBusy := busySlots[calSlot]
		if !isStillBusy {
			if err := s.cal.DeleteEvent(eventID); err != nil {
				return err
			}
		}
	}
	return nil
}
