package cron

import (
	"fmt"
	"time"

	"github.com/TBD54566975/ftl/internal/slices"
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

type componentType int

const (
	second componentType = iota
	minute
	hour
	dayOfMonth
	month     // 1 is Jan, 12 is Dec (same as time.Month)
	dayOfWeek // 0 and 7 are both Sunday (same as time.Weekday, except extra case of 7 == Sunday)
	year
)

// dayBehavior represents the behaviour of a cron pattern regarding which of the dayOfMonth and dayOfWeek components are used
type dayBehavior int

const (
	dayOfMonthOnly dayBehavior = iota
	dayOfWeekOnly
	dayOfMonthOrWeek
)

// componentValues represents the values of a time.Time in the order of componentType
// dayOfWeek is ignored
// a value of -1 represents a value that is not set (behaves as "lower than min value")
type componentValues []int

// Next calculates the next time that matches the pattern after the current time
// See NextAfter for more details
func Next(pattern Pattern, allowCurrentTime bool) (time.Time, error) {
	return NextAfter(pattern, time.Now().UTC(), allowCurrentTime)
}

// Next calculcates the next time that matches the pattern after the origin time
// If inclusive is true, the origin time is considered a valid match
// All calculations are done in UTC, and the result is returned in UTC
func NextAfter(pattern Pattern, origin time.Time, inclusive bool) (time.Time, error) {
	// set original to the first acceptable time, irregardless of pattern
	origin = origin.UTC()
	if !inclusive || origin.Nanosecond() != 0 {
		origin = origin.Add(time.Second - time.Duration(origin.Nanosecond())*time.Nanosecond)
	}

	components, err := pattern.standardizedComponents()
	if err != nil {
		return origin, err
	}

	for idx, component := range components {
		if err = validateComponent(component, componentType(idx)); err != nil {
			return origin, err
		}
	}

	// dayOfMonth used to represent processing day, using dayOfMonth and dayOfWeek
	processingOrder := []componentType{year, month, dayOfMonth, hour, minute, second}

	values := componentValuesFromTime(origin)

	firstDisallowedIdx := -1
	for idx, t := range processingOrder {
		if !isCurrentValueAllowed(components, values, t) {
			firstDisallowedIdx = idx
			break
		}
	}
	if firstDisallowedIdx == -1 {
		return timeFromValues(values), nil
	}

	i := firstDisallowedIdx
	for i >= 0 {
		t := processingOrder[i]
		next, err := nextValue(components, values, t)
		if err != nil {
			// no next value for this type, need to go up a level
			for ii := i; ii < len(processingOrder); ii++ {
				tt := processingOrder[ii]
				values[tt] = -1
			}
			i--
			continue
		}

		values[t] = next
		couldNotFindValueForIdx := -1
		for ii := i + 1; ii < len(processingOrder); ii++ {
			tt := processingOrder[ii]
			first, error := firstValueForComponents(components, values, tt)
			if error != nil {
				couldNotFindValueForIdx = ii
				break
			}
			values[tt] = first
		}
		if couldNotFindValueForIdx != -1 {
			// Could not find a value for a smaller type. Go up one level from that type
			i = couldNotFindValueForIdx - 1
			continue
		}

		return timeFromValues(values), nil
	}

	return origin, fmt.Errorf("could not find next time for pattern %q", pattern.String())
}

func componentValuesFromTime(t time.Time) componentValues {
	return []int{
		t.Second(),
		t.Minute(),
		t.Hour(),
		t.Day(),
		int(t.Month()),
		int(t.Weekday()),
		t.Year(),
	}
}

func isCurrentValueAllowed(components []Component, values componentValues, t componentType) bool {
	if t == dayOfWeek {
		// use dayOfMonth to check day of month and week
		panic("unexpected dayOfWeek value")
	} else if t == dayOfMonth {
		behavior := dayBehaviorForComponents(components)

		if behavior == dayOfMonthOnly || behavior == dayOfMonthOrWeek {
			if isCurrentValueAllowedForSteps(components[t].List, values, t) {
				return true
			}
		}
		if behavior == dayOfWeekOnly || behavior == dayOfMonthOrWeek {
			for _, step := range components[dayOfWeek].List {
				if isCurrentValueAllowedForDayOfWeekStep(step, values, t) {
					return true
				}
			}
		}
		return false
	}
	return isCurrentValueAllowedForSteps(components[t].List, values, t)
}

func isCurrentValueAllowedForSteps(steps []Step, values componentValues, t componentType) bool {
	for _, step := range steps {
		if isCurrentValueAllowedForStep(step, values, t) {
			return true
		}
	}
	return false
}

func isCurrentValueAllowedForStep(step Step, values componentValues, t componentType) bool {
	start, end, incr := rangeParametersForStep(step, t)
	if values[t] < start || values[t] > end {
		return false
	}
	if (values[t]-start)%incr != 0 {
		return false
	}
	return true
}

func isCurrentValueAllowedForDayOfWeekStep(step Step, values componentValues, t componentType) bool {
	start, end, incr := rangeParametersForStep(step, t)
	value := int(time.Date(values[year], time.Month(values[month]), values[dayOfMonth], 0, 0, 0, 0, time.UTC).Weekday())
	// Sunday is both 0 and 7
	days := []int{value}
	if value == 0 {
		days = append(days, 7)
	} else if value == 7 {
		days = append(days, 0)
	}

	results := slices.Map(days, func(day int) bool {
		if values[t] < start || values[t] > end {
			return false
		}
		if (values[t]-start)%incr != 0 {
			return false
		}
		return true
	})

	for _, result := range results {
		if result {
			return true
		}
	}
	return false
}

func nextValue(components []Component, values componentValues, t componentType) (int, error) {
	if t == dayOfWeek {
		// use dayOfMonth to check day of month and week
		panic("unexpected dayOfWeek value")
	} else if t == dayOfMonth {
		behavior := dayBehaviorForComponents(components)

		next := -1
		if behavior == dayOfMonthOnly || behavior == dayOfMonthOrWeek {
			if possible, err := nextValueForSteps(components[t].List, values, t); err == nil {
				if next == -1 || possible < next {
					next = possible
				}
			}
		}
		if behavior == dayOfWeekOnly || behavior == dayOfMonthOrWeek {
			for _, step := range components[dayOfWeek].List {
				if possible, ok := nextValueForDayOfWeekStep(step, values, t); ok {
					if next == -1 || possible < next {
						next = possible
					}
				}
			}
		}
		if next == -1 {
			return -1, fmt.Errorf("no next value for %s", stringForComponentType(t))
		}
		return next, nil
	}
	return nextValueForSteps(components[t].List, values, t)
}

func nextValueForSteps(steps []Step, values componentValues, t componentType) (int, error) {
	next := -1
	for _, step := range steps {
		if v, ok := nextValueForStep(step, values, t); ok {
			if next == -1 || v < next {
				next = v
			}
		}
	}
	if next == -1 {
		return -1, fmt.Errorf("no next value for %s", stringForComponentType(t))
	}
	return next, nil
}

func nextValueForStep(step Step, values componentValues, t componentType) (int, bool) {
	// Value of -1 means no existing value and the first valid value should be returned
	if t == dayOfWeek {
		// use dayOfMonth to check day of month and week
		panic("unexpected dayOfWeek value")
	}

	start, end, incr := rangeParametersForStep(step, t)

	current := values[t]
	next := -1
	if current < start {
		next = start
	} else {
		// round down to the nearest increment from start, then add one increment
		next = start + (((current-start)/incr)+1)*incr
	}
	if next < start || next > end {
		return -1, false
	}

	// Any type specific checks
	if t == dayOfMonth {
		date := time.Date(values[year], time.Month(values[month]), next, 0, 0, 0, 0, time.UTC)
		if date.Day() != next {
			// This month does not not have this day in this particular year (eg Feb 30th)
			return -1, false
		}
	}
	return next, true
}

func nextValueForDayOfWeekStep(step Step, values componentValues, t componentType) (int, bool) {
	start, end, incr := rangeParametersForStep(step, t)
	stepAllowsSecondSunday := (start <= 7 && end >= 7 && (7-start)%incr == 0)

	result := -1
	if standardResult, ok := nextValueForDayOfStandardizedWeekStep(step, values, t); ok {
		result = standardResult
	}
	// If Sunday as a value of 7 is allowed by step, check the logic with a value of 0
	if stepAllowsSecondSunday {
		if secondSundayResult, ok := nextValueForDayOfStandardizedWeekStep(newStepWithValue(0), values, t); ok {
			if result == -1 || secondSundayResult < result {
				result = secondSundayResult
			}
		}
	}
	return result, result != -1
}

func nextValueForDayOfStandardizedWeekStep(step Step, values componentValues, t componentType) (int, bool) {
	// Ignores Sunday == 7
	start, end, incr := rangeParametersForStep(step, t)
	if start == 7 {
		return -1, false
	}
	if end == 7 {
		end = 6
	}

	valueForCurrentWeekday := max(0, values[dayOfMonth]) // is value is -1, we want day before the current month (ie 0)
	currentDate := time.Date(values[year], time.Month(values[month]), valueForCurrentWeekday, 0, 0, 0, 0, time.UTC)
	currentWeekday := int(currentDate.Weekday())

	startOfWeekInMonth := valueForCurrentWeekday - currentWeekday //Sunday

	nextDayOfWeek := -1

	// try current week
	if currentWeekday < start {
		nextDayOfWeek = start
	} else {
		// round down to the nearest increment from start, then add one increment
		nextDayOfWeek = start + (((currentWeekday-start)/incr)+1)*incr
	}
	if nextDayOfWeek < start || nextDayOfWeek > end {
		// try next week
		nextDayOfWeek = 7 + start
	}

	next := startOfWeekInMonth + nextDayOfWeek

	date := time.Date(values[year], time.Month(values[month]), next, 0, 0, 0, 0, time.UTC)
	if date.Day() != next {
		// This month does not not have this day in this particular year (eg Feb 30th)
		return -1, false
	}
	return next, true
}

func firstValueForComponents(components []Component, values componentValues, t componentType) (int, error) {
	fakeValues := make([]int, len(values))
	copy(fakeValues, values)
	fakeValues[t] = -1
	return nextValue(components, fakeValues, t)
}

func timeFromValues(values componentValues) time.Time {
	return time.Date(values[year],
		time.Month(values[month]),
		values[dayOfMonth],
		values[hour],
		values[minute],
		values[second],
		0,
		time.UTC)
}

// Validite makes sure that a pattern has no mistakes in the cron format, and that there is a valid next value from a set point in time
// Validity checks are done while calculating a next date to ensure that we never calculate a next date for an invalid pattern
func Validate(pattern Pattern) error {
	_, err := NextAfter(pattern, time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC), true)
	return err
}

func validateComponent(component Component, t componentType) error {
	if len(component.List) == 0 {
		return fmt.Errorf("%s must have at least value", stringForComponentType(t))
	}
	for _, step := range component.List {
		if step.ValueRange.IsFullRange && (step.ValueRange.Start != nil || step.ValueRange.End != nil) {
			return fmt.Errorf("range can not have start/end if it is a full range")
		}
		if !step.ValueRange.IsFullRange {
			min, max := rangeForComponentType(t)

			if step.ValueRange.Start == nil {
				return fmt.Errorf("missing value in %s", stringForComponentType(t))
			}
			if *step.ValueRange.Start < min || *step.ValueRange.Start > max {
				return fmt.Errorf("value %d out of allowed %s range of %d-%d", *step.ValueRange.Start, stringForComponentType(t), min, max)
			}
			if step.ValueRange.End != nil {
				if *step.ValueRange.End < min || *step.ValueRange.End > max {
					return fmt.Errorf("value %d out of allowed %s range of %d-%d", *step.ValueRange.End, stringForComponentType(t), min, max)
				}
				if *step.ValueRange.End < *step.ValueRange.Start {
					return fmt.Errorf("range end %d is less than start %d", *step.ValueRange.End, *step.ValueRange.Start)
				}
			}

			if step.Step != nil {
				if *step.Step <= 0 {
					return fmt.Errorf("step must be positive")
				}
				if *step.Step > max-min {
					return fmt.Errorf("step %d is larger than allowed range of %d-%d", *step.Step, max, min)
				}
				if t == year && step.ValueRange.IsFullRange {
					// This may be supported in other cron implementations, but will require more research as to the correct behavior
					return fmt.Errorf("asterix with a step value is not allowed for year component")
				}
			}
		}
	}

	return nil
}

func rangeForComponentType(t componentType) (min int, max int) {
	switch t {
	case second, minute:
		return 0, 59
	case hour:
		return 0, 23
	case dayOfMonth:
		return 1, 31
	case month:
		return 1, 12
	case dayOfWeek:
		return 0, 7
	case year:
		return 0, 3000
	default:
		panic("unknown component type")
	}
}

func rangeParametersForStep(step Step, t componentType) (start, end, incr int) {
	start, end = rangeForComponentType(t)
	incr = 1
	if step.Step != nil {
		incr = *step.Step
	}
	if step.ValueRange.Start != nil {
		start = *step.ValueRange.Start
		if step.ValueRange.End == nil {
			// "1/2" means start at 1 and increment by 2
			// "1" means "1-1"
			if step.Step == nil {
				end = start
			}
		} else {
			end = *step.ValueRange.End
		}
	}
	return
}

func dayBehaviorForComponents(components []Component) dayBehavior {
	// Spec: https://pubs.opengroup.org/onlinepubs/9699919799.2018edition/utilities/crontab.html
	isMonthAsterix := components[month].String() == "*"
	isDayOfMonthAsterix := components[dayOfMonth].String() == "*"
	isDayOfWeekAsterix := components[dayOfWeek].String() == "*"

	// If month, day of month, and day of week are all <asterisk> characters, every day shall be matched.
	if isMonthAsterix && isDayOfMonthAsterix && isDayOfWeekAsterix {
		return dayOfMonthOnly
	}

	// If either the month or day of month is specified as an element or list, but the day of week is an <asterisk>, the month and day of month fields shall specify the days that match.
	if (!isMonthAsterix || !isDayOfMonthAsterix) && isDayOfWeekAsterix {
		return dayOfMonthOnly
	}

	// If both month and day of month are specified as an <asterisk>, but day of week is an element or list, then only the specified days of the week match.
	if isMonthAsterix && isDayOfMonthAsterix && !isDayOfWeekAsterix {
		return dayOfWeekOnly
	}

	// Finally, if either the month or day of month is specified as an element or list, and the day of week is also specified as an element or list, then any day matching either the month and day of month, or the day of week, shall be matched.
	return dayOfMonthOrWeek
}

func stringForComponentType(t componentType) string {
	switch t {
	case second:
		return "second"
	case minute:
		return "minute"
	case hour:
		return "hour"
	case dayOfMonth:
		return "day of month"
	case month:
		return "month"
	case dayOfWeek:
		return "day of week"
	case year:
		return "year"
	default:
		panic("unknown component type")
	}
}
