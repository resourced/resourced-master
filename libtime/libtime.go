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

func IsLeapYear(y int) bool {
	year := time.Date(y, time.December, 31, 0, 0, 0, 0, time.Local)
	days := year.YearDay()

	if days > 365 {
		return true
	}

	return false
}
