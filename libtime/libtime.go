// Package libtime provides time related library functions.
package libtime

import (
	"time"
)

// SleepString is a convenience function that performs `time.Sleep` given string duration.
func SleepString(definition string) error {
	delayTime, err := time.ParseDuration(definition)
	if err != nil {
		return err
	}

	time.Sleep(delayTime)
	return nil
}

// IsLeapYear check if given year is a leap year.
func IsLeapYear(y int) bool {
	year := time.Date(y, time.December, 31, 0, 0, 0, 0, time.Local)
	days := year.YearDay()

	if days > 365 {
		return true
	}

	return false
}

func FirstOfNextMonthUTC() time.Time {
	now := time.Now().UTC()
	currentYear, currentMonth, _ := now.Date()

	nextMonth := currentMonth + 1
	if nextMonth > 12 {
		nextMonth = nextMonth - 12
	}

	return time.Date(currentYear, nextMonth, 1, 0, 0, 0, 0, now.Location()).UTC()
}
