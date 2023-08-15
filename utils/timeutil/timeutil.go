package timeutil

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
	"ytc/defs/regexdef"
)

const (
	YEAR_INDEX = iota
	MONTH_INDEX
	DAY_INDEX
	HOUR_INDEX
	MINUTE_INDEX
	SECOND_INDEX
)

const (
	minute = "m"
	hour   = "h"
	day    = "d"
	month  = "M"
	year   = "y"

	minute_dur time.Duration = time.Minute
	hour_dur                 = time.Hour
	day_dur                  = hour_dur * 24
	month_dur                = day_dur * 30
	year_dur                 = month_dur * 12
)

var (
	ErrDurationInvalid = errors.New("duration invalid")
)

func GetTimeDivBySepa(timeStr, sepa string) (time.Time, error) {
	dateFields := [6]int{}
	parts := strings.Split(timeStr, sepa)
	for index := range parts {
		field, err := strconv.ParseInt(parts[index], 10, 64)
		if err != nil {
			return time.Time{}, fmt.Errorf("prase time err: %s", err.Error())
		}
		dateFields[index] = int(field)
	}
	t := time.Date(
		dateFields[YEAR_INDEX],
		time.Month(dateFields[MONTH_INDEX]),
		dateFields[DAY_INDEX],
		dateFields[HOUR_INDEX],
		dateFields[MINUTE_INDEX],
		dateFields[SECOND_INDEX],
		0, time.Local,
	)
	return t, nil
}

func GetDuration(s string) (d time.Duration, err error) {
	if !regexdef.RANGE_REGEX.MatchString(s) {
		err = ErrDurationInvalid
		return
	}
	var p int64
	var dunit time.Duration
	suffix := s[len(s)-1:]
	prefix := s[:len(s)-1]
	p, err = strconv.ParseInt(prefix, 10, 64)
	if err != nil {
		return
	}
	switch suffix {
	case year:
		dunit = year_dur
	case month:
		dunit = month_dur
	case day:
		dunit = day_dur
	case hour:
		dunit = hour_dur
	case minute:
		dunit = minute_dur
	}
	d = duration(p, dunit)
	return
}

func duration(base int64, unit time.Duration) time.Duration {
	return time.Duration(base * int64(unit))
}
