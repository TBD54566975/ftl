package cron

import (
	"fmt"
	"time"
)

/*
 This cron package is a simple implementation of a cron pattern parser and evaluator.
 It supports the following:
 - 5 component patterns interpreted as second, minute, hour, day of month, month
 - 6 component patterns interpreted as:
   - if last component has a 4 digit number, it is interpreted as minute, hour, day of month, month, year
   - otherwise, it is interpreted as second, minute, hour, day of month, month, day of week
- 7 component patterns, interpreted as second, minute, hour, day of month, month, day of week, year

It supports the following features:
- * for all values
- ranges with - (eg 1-5)
- steps with / (eg 1-5/2)
- lists with , (eg 1,2,3)
*/

// Next calculates the next time that matches the pattern after the current time
// See NextAfter for more details
func Next(pattern Pattern, allowCurrentTime bool) (time.Time, error) {
	return NextAfter(pattern, time.Now().UTC(), allowCurrentTime)
}

// NextAfter calculates the next time that matches the pattern after the origin time
// If inclusive is true, the origin time is considered a valid match
// All calculations are done in UTC, and the result is returned in UTC
func NextAfter(pattern Pattern, origin time.Time, inclusive bool) (time.Time, error) {
	// set original to the first acceptable time, regardless of pattern
	origin = origin.UTC()
	if inclusive && origin.Unix() != 0 {
		origin = origin.Add(-time.Second)
	}

	next := pattern.expression.Next(origin)
	if next.IsZero() {
		return next, fmt.Errorf("unable to find next timeout for %s", pattern)
	}
	return next, nil
}
