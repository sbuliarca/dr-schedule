package ds

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"google.golang.org/api/calendar/v3"
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
	cal          *Cal
	portalClient PortalClient
}

func NewService(
	cal *Cal,
	portalClient PortalClient,
) *Service {
	return &Service{
		cal:          cal,
		portalClient: portalClient,
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
const daysToSchedule = 7

// todo parameterize the schedule
var Schedule = map[time.Weekday]DaySchedule{
	time.Monday: {
		StartHour: 17, StartMin: 30,
		EndHour: 20, EndMin: 0,
		SlotDuration: 30,
	},
	time.Tuesday: {
		StartHour: 8, StartMin: 0,
		EndHour: 10, EndMin: 30,
		SlotDuration: 30,
	},
	time.Wednesday: {
		StartHour: 17, StartMin: 30,
		EndHour: 20, EndMin: 0,
		SlotDuration: 30,
	},
	time.Thursday: {
		StartHour: 8, StartMin: 0,
		EndHour: 10, EndMin: 30,
		SlotDuration: 30,
	},
	time.Friday: {
		StartHour: 17, StartMin: 30,
		EndHour: 20, EndMin: 0,
		SlotDuration: 30,
	},
}

func (s *Service) SyncSlots(startTime time.Time) error {
	logrus.Infof("Started syncing slots with start time %s", startTime.Format(time.RFC3339))
	startDate := time.Date(startTime.Year(), startTime.Month(), startTime.Day(), 0, 0, 0, 0, LocalLoc)

	freeSlots, err := s.portalClient.GetFreeSlots(startDate, daysToSchedule)
	logrus.Debugf("found free slots %v", freeSlots)
	if err != nil {
		return err
	}

	calEvents, err := s.cal.GetEvents(startDate, daysToSchedule)
	if err != nil {
		return fmt.Errorf("failed getting the current calendar events, err: %w", err)
	}
	logDebugEvents(calEvents)

	for i := 0; i < daysToSchedule; i++ {
		dayToProcess := startDate.AddDate(0, 0, i)
		day := dayToProcess.Weekday()

		daySched, hasSchedule := Schedule[day]
		if !hasSchedule {
			continue
		}
		freeDaySlots := getFreeDaySlots(freeSlots, dayToProcess)
		logrus.Debugf("found free slots for day %v : %v", dayToProcess, freeDaySlots)

		if err := s.syncDay(dayToProcess, daySched, freeDaySlots, calEvents); err != nil {
			return err
		}
	}

	logrus.Info("Finished syncing slots")
	return nil
}

func logDebugEvents(calEvents *calendar.Events) {
	var evs []calendar.Event
	logrus.Debugf("Found existing events: %v", evs)

	for _, cali := range calEvents.Items {
		evs = append(evs, *cali)
	}
}

func (s *Service) syncDay(day time.Time, sched DaySchedule, slots *FreeDaySlots, calEvents *calendar.Events) error {
	slotTime := time.Date(day.Year(), day.Month(), day.Day(), sched.StartHour, sched.StartMin, 0, 0, LocalLoc)
	endTime := time.Date(day.Year(), day.Month(), day.Day(), sched.EndHour, sched.EndMin, 0, 0, LocalLoc)

	for slotTime.Before(endTime) {
		slotIsFree := isSlotFree(slotTime, slots)
		hasCalEvent, calEventId, err := hasCalendarEvent(slotTime, calEvents)
		if err != nil {
			return fmt.Errorf("failed determining if calendar has event %w", err)
		}
		logrus.Debugf("slot at time %v is free %v and has calendar event: %v", slotTime, slotIsFree, hasCalEvent)

		switch {
		case !slotIsFree && !hasCalEvent: // create the event if it doesn't exist already
			if err := s.cal.CreateEvent(slotTime, sched.SlotDuration); err != nil {
				return err
			}
		case slotIsFree && hasCalEvent: // delete event if the slot is free
			if err := s.cal.DeleteEvent(calEventId); err != nil {
				return err
			}
		}

		slotTime = slotTime.Add(time.Duration(sched.SlotDuration) * time.Minute)
	}
	return nil
}

func hasCalendarEvent(slotTime time.Time, events *calendar.Events) (bool, string, error) {
	for _, calEv := range events.Items {
		evStart, err := time.Parse(time.RFC3339, calEv.Start.DateTime)
		if err != nil {
			return false, "", err
		}
		if calEv.Description == EventDescription && evStart.Equal(slotTime) {
			logrus.Debugf("found existing event for slot time %s : %v", slotTime.Format(time.RFC3339), *calEv)
			return true, calEv.Id, nil
		}
	}
	return false, "", nil
}

func isSlotFree(slotTime time.Time, slots *FreeDaySlots) bool {
	if slots == nil {
		return false
	}
	timeStr := slotTime.Format("15:04")
	return strings.Contains(slots.Hours, timeStr)
}

func getFreeDaySlots(slots []FreeDaySlots, day time.Time) *FreeDaySlots {
	dayStr := day.Format(DateFormat)
	for _, slot := range slots {
		if slot.Date == dayStr {
			return &slot
		}
	}
	return nil
}
