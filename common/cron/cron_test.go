package cron

import (
	"fmt"
	"testing"
	"time"

	"github.com/alecthomas/assert/v2"
)

func TestNonUTC(t *testing.T) {
	// This cron package only works with UTC times.
	// Passing in non-UTC times works fine, but the results will be in UTC.
}

func TestParsingAndValidationErrors(t *testing.T) {
	// Rather than testing successful parsing, test them in TestNext()
	for _, tt := range []struct {
		str string
		err string
	}{
		{"* * * *", "expected 5-7 components, got 4"},
		{"* * * * * * * *", "expected 5-7 components, got 8"},
		{"1-10,4-5/1,59-61 * * * *", "value 61 out of allowed minute range of 0-59"},
		{"4-5 * * 13 *", "value 13 out of allowed month range of 1-12"},
		{"4-5 * * -1 *", "1:9: unexpected token \"-\""},
		{"4-5 * * 0 *", "value 0 out of allowed month range of 1-12"},
		{"* * * * * 9999", "value 9999 out of allowed year range of 0-3000"},
		{"* * 30 2 *", "could not find next time for pattern \"* * 30 2 *\""},
		{"* * 30/0 * *", "step must be positive"},
		{"* * * * * 1999", "could not find next time for pattern \"* * * * * 1999\""},
		{"* * * * * * 1999", "could not find next time for pattern \"* * * * * * 1999\""},
		{"* * * 29 2 * 2021", "could not find next time for pattern \"* * * 29 2 * 2021\""},
	} {
		t.Run(fmt.Sprintf("CronValidation:%s", tt.str), func(t *testing.T) {
			_, err := Parse(tt.str)
			assert.Error(t, err)
		})
	}
}

func TestNext(t *testing.T) {
	//TODO: test inputting non UTC...
	for _, tt := range []struct {
		str              string
		inputsAndOutputs [][]time.Time
	}{
		{"* * * * * * *", [][]time.Time{
			{
				time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
				time.Date(2020, 1, 1, 0, 0, 1, 0, time.UTC),
			},
			{ // Ticking over midnight
				time.Date(2020, 1, 10, 23, 59, 59, 546, time.UTC),
				time.Date(2020, 1, 11, 0, 0, 0, 0, time.UTC),
			},
			{ // Ticking over midnight at the end of feb, not on a leap year
				time.Date(2022, 2, 28, 23, 59, 59, 666, time.UTC),
				time.Date(2022, 3, 1, 0, 0, 0, 0, time.UTC),
			},
			{ // Ticking over midnight at the end of feb, on a leap year
				time.Date(2024, 2, 28, 23, 59, 59, 666, time.UTC),
				time.Date(2024, 2, 29, 0, 0, 0, 0, time.UTC),
			},
		}},
		{"* * * * *", [][]time.Time{
			{
				time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
				time.Date(2020, 1, 1, 0, 1, 0, 0, time.UTC),
			},
			{
				time.Date(2020, 1, 19, 3, 34, 0, 234, time.UTC),
				time.Date(2020, 1, 19, 3, 35, 0, 0, time.UTC),
			},
			{ // A minute over an hour
				time.Date(2020, 1, 19, 3, 59, 0, 234, time.UTC),
				time.Date(2020, 1, 19, 4, 0, 0, 0, time.UTC),
			},
			{ // A minute over midnight
				time.Date(2020, 1, 10, 23, 59, 3, 546, time.UTC),
				time.Date(2020, 1, 11, 0, 0, 0, 0, time.UTC),
			},
			{ // A minute over midnight at the end of feb, not on a leap year
				time.Date(2022, 2, 28, 23, 59, 6, 666, time.UTC),
				time.Date(2022, 3, 1, 0, 0, 0, 0, time.UTC),
			},
			{ // A minute over midnight at the end of feb, on a leap year
				time.Date(2024, 2, 28, 23, 59, 55, 666, time.UTC),
				time.Date(2024, 2, 29, 0, 0, 0, 0, time.UTC),
			},
		}},
		// 6 components, should be treated as "every 10 seconds:
		{"*/10 * * * * *", [][]time.Time{
			{
				time.Date(2020, 1, 1, 0, 0, 17, 0, time.UTC),
				time.Date(2020, 1, 1, 0, 0, 20, 0, time.UTC),
			},
		}},
		// 6 components, should be treated as "every 10 minutes, every second year"
		{"*/10 * * * * 2022/2", [][]time.Time{
			{
				time.Date(2023, 6, 9, 18, 12, 2, 300, time.UTC),
				time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			{
				time.Date(2024, 6, 9, 18, 12, 2, 300, time.UTC),
				time.Date(2024, 6, 9, 18, 20, 0, 0, time.UTC),
			},
		}},
		// */5 * * * * * *
		{"5s", [][]time.Time{
			{
				time.Date(2025, 6, 5, 3, 7, 5, 123, time.UTC),
				time.Date(2025, 6, 5, 3, 7, 10, 0, time.UTC),
			},
			{
				time.Date(2025, 6, 5, 3, 59, 55, 123, time.UTC),
				time.Date(2025, 6, 5, 4, 0, 0, 0, time.UTC),
			},
		}},
		// 25m should be every 25 minutes: 0 */25 * * * * * ie 0,25,50
		{"25m", [][]time.Time{
			{
				time.Date(2025, 6, 5, 3, 7, 5, 123, time.UTC),
				time.Date(2025, 6, 5, 3, 25, 0, 0, time.UTC),
			},
			{
				time.Date(2025, 6, 5, 3, 49, 5, 123, time.UTC),
				time.Date(2025, 6, 5, 3, 50, 0, 0, time.UTC),
			},
			{
				time.Date(2025, 6, 5, 3, 50, 5, 123, time.UTC),
				time.Date(2025, 6, 5, 4, 0, 0, 0, time.UTC),
			},
		}},
		// 5h should be every 5 hours: 0 0 */5 * * * *, ie 0,5,10,15,20
		{"5h", [][]time.Time{
			{
				time.Date(2025, 6, 5, 3, 7, 5, 123, time.UTC),
				time.Date(2025, 6, 5, 5, 0, 0, 0, time.UTC),
			},
			{
				time.Date(2025, 6, 5, 19, 59, 5, 123, time.UTC),
				time.Date(2025, 6, 5, 20, 0, 0, 0, time.UTC),
			},
			{
				time.Date(2025, 6, 5, 21, 59, 5, 123, time.UTC),
				time.Date(2025, 6, 6, 0, 0, 0, 0, time.UTC),
			},
		}},
		// Every wednesday
		{"0 0 0 * * 3 *", [][]time.Time{
			{ // 2024-06-09 is a Sunday
				time.Date(2024, 6, 9, 0, 0, 0, 0, time.UTC),
				time.Date(2024, 6, 12, 0, 0, 0, 0, time.UTC),
			},
		}},
		{"Wed", [][]time.Time{
			{ // 2024-06-09 is a Sunday
				time.Date(2024, 6, 9, 0, 0, 0, 0, time.UTC),
				time.Date(2024, 6, 12, 0, 0, 0, 0, time.UTC),
			},
		}},
	} {
		t.Run(fmt.Sprintf("CronSeries:%s", tt.str), func(t *testing.T) {
			pattern, err := Parse(tt.str)
			assert.NoError(t, err)
			for _, inputAndOutput := range tt.inputsAndOutputs {
				input := inputAndOutput[0]
				output := inputAndOutput[1]
				next, err := NextAfter(pattern, input, false)
				assert.NoError(t, err)
				assert.Equal(t, output, next, "NextAfter(%q, %v) = %v; want %v", tt.str, input, next, output)

				outputAsInput, err := NextAfter(pattern, output, true)
				assert.NoError(t, err)
				assert.Equal(t, outputAsInput, output, "output of Next() should also satisfy NextAfter() with inclusive=true")
			}
		})
	}
}

func TestSeries(t *testing.T) {
	for _, tt := range []struct {
		str           string
		input         time.Time
		end           time.Time
		expectedCount int
	}{
		{
			"* * * * * * *",
			time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			time.Date(2020, 1, 1, 0, 0, 10, 0, time.UTC),
			10,
		},
		{
			"* * * * * * *",
			time.Date(2020, 1, 1, 0, 0, 50, 0, time.UTC),
			time.Date(2020, 1, 1, 0, 1, 10, 0, time.UTC),
			20,
		},
		{ // Every 31st of the month in a year
			"0 0 0 31 * * *",
			time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
			7,
		},
		{ // Every 29th of Feb in the 2020s
			"0 0 0 29 2 * *",
			time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC),
			3,
		},
		{ // Five Mondays in Jan 2024
			"0 0 0 * * 1 *",
			time.Date(2023, 12, 31, 23, 59, 0, 0, time.UTC),
			time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC),
			5,
		},
		{ // Four Sundays in Jan 2024 (Sunday == 0)
			"0 0 0 * * 0 *",
			time.Date(2023, 12, 31, 23, 59, 0, 0, time.UTC),
			time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC),
			4,
		},
		{ // Four Sundays in Jan 2024 (sunday == 7)
			"0 0 0 * * 7 *",
			time.Date(2023, 12, 31, 23, 59, 0, 0, time.UTC),
			time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC),
			4,
		},
		{ // Each Mon/Wed/Friday/Sun in Jan 2024
			"0 0 0 * * 1/2 *",
			time.Date(2023, 12, 31, 23, 59, 0, 0, time.UTC),
			time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC),
			14,
		},
		{ // 10,11,12,13,14,17,19,24,36,48
			"12/12,10-14,17-20/2 * * * * * *",
			time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			time.Date(2024, 1, 1, 0, 0, 59, 100, time.UTC),
			10,
		},
		{ // Each Mon/Wed/Friday/Sun, AND the 9th in Jan 2024
			"0 0 0 9 * 1/2 *",
			time.Date(2023, 12, 31, 23, 59, 0, 0, time.UTC),
			time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC),
			15,
		},
		{ // Each Mon/Wed/Friday/Sun, AND the 8th (which is a Monday anyway) in Jan 2024
			"0 0 0 8 * 1/2 *",
			time.Date(2023, 12, 31, 23, 59, 0, 0, time.UTC),
			time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC),
			14,
		},
		{ // Each Mon/Wed/Friday/Sun, AND every day of Jan in Jan 2024
			"0 0 0 * 1 1/2 *",
			time.Date(2023, 12, 31, 23, 59, 0, 0, time.UTC),
			time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC),
			14,
		},
		{ // An hour worth of 5 minutes
			"0 */5 * * * * *",
			time.Date(2025, 1, 2, 3, 4, 5, 6, time.UTC),
			time.Date(2025, 1, 2, 4, 4, 5, 6, time.UTC),
			12,
		},
		{ // An hour worth of 5 minutes using shorthand
			"5m",
			time.Date(2025, 1, 2, 3, 4, 5, 6, time.UTC),
			time.Date(2025, 1, 2, 4, 4, 5, 6, time.UTC),
			12,
		},
		{ // A month of Fridays
			"0 0 0 * * 5 *",
			// 2025-01-01 is a Wednesday
			time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC),
			5,
		},
		{ // A month of Fridays
			"fri",
			// 2025-01-01 is a Wednesday
			time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC),
			5,
		},
	} {
		t.Run(fmt.Sprintf("CronSeries:%s", tt.str), func(t *testing.T) {
			pattern, err := Parse(tt.str)
			assert.NoError(t, err)

			value, err := NextAfter(pattern, tt.input, false)
			assert.NoError(t, err)

			count := 0
			for !value.After(tt.end) {
				count++

				value, err = NextAfter(pattern, value, false)
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.expectedCount, count, "Count of %q between %v - %v) = %v; want %v", tt.str, tt.input, tt.end, count, tt.expectedCount)
		})
	}
}
