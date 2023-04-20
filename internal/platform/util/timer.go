package util

import (
	"fmt"
	"log"
	"time"
)

const (
	superSmallLayout     = "Jan 2"
	smallLayout          = "Jan 2 2006, 3:04PM"
	dateTimeGoLayout     = "2006-01-02 15:04:05 -0700"
	dateTimeGoogleLayout = "2006-01-02T15:04:05-07:00"
)

func GetMilliSeconds(now time.Time) int64 {
	return now.UTC().Unix() * 1000
}

func GetMilliSecondsFloat(now time.Time) float64 { // use this until redis g adds support for int64 in ToString method
	return float64(GetMilliSeconds(now))
}

func GetMilliSecondsFloatReduced(now time.Time) float64 { // use this until redis g adds support for int64 in ToString method
	return float64(GetMilliSeconds(now)) / 1000
}

func AddMilliSecondsFloat(now time.Time, days int) float64 { // use this until redis g adds support for int64 in ToString method
	return float64(GetMilliSecondsFloatReduced(now.AddDate(0, 0, days)))
}

func GetMilliSecondsStr(now time.Time) string {
	return fmt.Sprintf("%d", GetMilliSeconds(now))
}

func IsValidTime(fromTime time.Time) bool {
	return fromTime.Year() != 1
}

func Round(fromTime time.Time, hr int) time.Time {
	t := time.Date(fromTime.Year(), fromTime.Month(), fromTime.Day(), hr, 0, 0, 0, fromTime.Location())
	return t
}

func ParseTime(str string) (time.Time, error) {
	return time.Parse(dateTimeGoLayout, str)
}

func FormatTimeGo(t time.Time) string {
	return t.Format(dateTimeGoLayout)
}

func FormatTimeGoogle(t time.Time) string {
	return t.Format(dateTimeGoogleLayout)
}

func FormatTimeView(t time.Time) string {
	return t.Format(smallLayout)
}

func FormatTimeViewSmall(t time.Time) string {
	return t.Format(superSmallLayout)
}

func ConvertMillisToTime(millis string) time.Time {
	millsL := ConvertStrToInt64(millis)
	return time.Unix(0, millsL*int64(time.Millisecond))
}

func ConvertMilliToTime(tm int64) time.Time {
	sec := tm / 1000
	msec := tm % 1000
	return time.Unix(sec, msec*int64(time.Millisecond))
}

func ConvertMilliToTimeFromIntf(val interface{}) string {
	var millis int64
	switch t := val.(type) {
	case int:
		millis = int64(t)
	case int32:
		millis = int64(t)
	case int64:
		millis = t
	case float64:
		millis = int64(t)
	case string:
		millis = ConvertStrToInt64(t)
	default:
		log.Printf("val THREE %+v", t)
	}
	return FormatTimeGoogle(time.Unix(0, millis*int64(time.Millisecond)))
}
