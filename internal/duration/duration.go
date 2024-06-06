package duration

import (
	"fmt"
	"regexp"
	"strconv"
	"time"
)

func ParseDuration(str string) (time.Duration, error) {
	// regex is more lenient than what is valid to allow for better error messages.
	re := regexp.MustCompile(`^(\d+)([a-zA-Z]+)`)

	var duration time.Duration
	previousUnitDuration := time.Duration(0)
	for len(str) > 0 {
		matches := re.FindStringSubmatchIndex(str)
		if matches == nil {
			return 0, fmt.Errorf("unable to parse duration %q - expected duration in format like '1m' or '30s'", str)
		}
		num, err := strconv.Atoi(str[matches[2]:matches[3]])
		if err != nil {
			return 0, fmt.Errorf("unable to parse duration %q: %w", str, err)
		}

		unitStr := str[matches[4]:matches[5]]
		var unitDuration time.Duration
		switch unitStr {
		case "d":
			unitDuration = time.Hour * 24
		case "h":
			unitDuration = time.Hour
		case "m":
			unitDuration = time.Minute
		case "s":
			unitDuration = time.Second
		default:
			return 0, fmt.Errorf("duration has unknown unit %q - use 'd', 'h', 'm' or 's', eg '1d' or '30s'", unitStr)
		}
		if previousUnitDuration != 0 && previousUnitDuration <= unitDuration {
			return 0, fmt.Errorf("duration has unit %q out of order - units need to be ordered from largest to smallest - eg '1d3h2m'", unitStr)
		}
		previousUnitDuration = unitDuration
		duration += time.Duration(num) * unitDuration
		str = str[matches[1]:]
	}

	return duration, nil
}
