package duration

import (
	"fmt"
	"regexp"
	"strconv"
	"time"
)

type Components struct {
	Days    int
	Hours   int
	Minutes int
	Seconds int
}

func (c Components) Duration() time.Duration {
	return time.Duration(c.Days*24)*time.Hour +
		time.Duration(c.Hours)*time.Hour +
		time.Duration(c.Minutes)*time.Minute +
		time.Duration(c.Seconds)*time.Second
}

func Parse(str string) (time.Duration, error) {
	components, err := ParseComponents(str)
	if err != nil {
		return 0, err
	}

	return components.Duration(), nil
}

func ParseComponents(str string) (*Components, error) {
	// regex is more lenient than what is valid to allow for better error messages.
	re := regexp.MustCompile(`^(\d+)([a-zA-Z]+)`)

	var components Components
	previousUnitDuration := time.Duration(0)
	for len(str) > 0 {
		matches := re.FindStringSubmatchIndex(str)
		if matches == nil {
			return nil, fmt.Errorf("unable to parse duration %q - expected duration in format like '1m' or '30s'", str)
		}
		num, err := strconv.Atoi(str[matches[2]:matches[3]])
		if err != nil {
			return nil, fmt.Errorf("unable to parse duration %q: %w", str, err)
		}

		unitStr := str[matches[4]:matches[5]]
		var unitDuration time.Duration
		switch unitStr {
		case "d":
			components.Days = num
			unitDuration = time.Hour * 24
		case "h":
			components.Hours = num
			unitDuration = time.Hour
		case "m":
			components.Minutes = num
			unitDuration = time.Minute
		case "s":
			components.Seconds = num
			unitDuration = time.Second
		default:
			return nil, fmt.Errorf("duration has unknown unit %q - use 'd', 'h', 'm' or 's', eg '1d' or '30s'", unitStr)
		}
		if previousUnitDuration != 0 && previousUnitDuration <= unitDuration {
			return nil, fmt.Errorf("duration has unit %q out of order - units need to be ordered from largest to smallest - eg '1d3h2m'", unitStr)
		}
		previousUnitDuration = unitDuration
		str = str[matches[1]:]
	}

	return &components, nil
}
