package util

import (
	"time"
)

const (
	smallLayout      = "Jan 2 2006, 3:04PM"
	dateTimeGoLayout = "2006-01-02 15:04:05 -0700"
)

func GetMilliSeconds(now time.Time) int64 {
	return now.UTC().Unix() * 1000
}

func IsValidTime(fromTime time.Time) bool {
	return fromTime.Year() != 1
}

func Round(fromTime time.Time, hr int) time.Time {
	t := time.Date(fromTime.Year(), fromTime.Month(), fromTime.Day(), hr, 0, 0, 0, fromTime.Location())
	return t
}

func FormatTimeGo(t time.Time) string {
	return t.Format(dateTimeGoLayout)
}

func ParseTime(str string) (time.Time, error) {
	return time.Parse(dateTimeGoLayout, str)
}