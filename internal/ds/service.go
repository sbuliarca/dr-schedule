package ds

import (
	"strings"
	"time"
)

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

const daysToSchedule = 7

var Schedule map[time.Weekday]DaySchedule

func (s *Service) SyncSlots(startTime time.Time) error {
	startDate := time.Date(startTime.Year(), startTime.Month(), startTime.Day(), 0, 0, 0, 0, time.UTC)

	freeSlots, err := s.portalClient.GetFreeSlots(startDate, daysToSchedule)
	if err != nil {
		return err
	}

	for i := 0; i < daysToSchedule; i++ {
		dayToProcess := startDate.AddDate(0, 0, i)
		day := dayToProcess.Weekday()

		daySched, hasSchedule := Schedule[day]
		if !hasSchedule {
			continue
		}
		freeDaySlots := getFreeDaySlots(freeSlots, dayToProcess)
		if freeDaySlots == nil {
			continue
		}

		if err := s.syncDay(dayToProcess, daySched, freeDaySlots); err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) syncDay(day time.Time, sched DaySchedule, slots *FreeDaySlots) error {
	slotTime := time.Date(day.Year(), day.Month(), day.Day(), sched.StartHour, sched.StartMin, 0, 0, time.UTC)
	endTime := time.Date(day.Year(), day.Month(), day.Day(), sched.EndHour, sched.EndMin, 0, 0, time.UTC)

	for slotTime.Before(endTime) {
		slotIsFree := isSlotFree(slotTime, slots)
		if !slotIsFree {
			if err := s.cal.CreateAppointment(slotTime, sched.SlotDuration); err != nil {
				return err
			}
		}

		slotTime = slotTime.Add(time.Duration(sched.SlotDuration) * time.Minute)
	}
	return nil
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
