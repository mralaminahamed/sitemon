package scheduler

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type Schedule struct {
	Expression string
	NextRun    time.Time
}

func ParseCron(expression string) (*Schedule, error) {
	parts := strings.Fields(expression)
	if len(parts) < 5 {
		return nil, fmt.Errorf("invalid cron expression: %s", expression)
	}

	return &Schedule{
		Expression: expression,
	}, nil
}

func (s *Schedule) Next() time.Time {
	return calculateNextRun(s.Expression, time.Now())
}

func calculateNextRun(expr string, from time.Time) time.Time {
	parts := strings.Fields(expr)

	minute, err := parseField(parts[0], 0, 59)
	if err != nil {
		return from
	}

	hour, err := parseField(parts[1], 0, 23)
	if err != nil {
		return from
	}

	dayOfMonth, err := parseField(parts[2], 1, 31)
	if err != nil {
		return from
	}

	month, err := parseField(parts[3], 1, 12)
	if err != nil {
		return from
	}

	dayOfWeek, err := parseField(parts[4], 0, 6)
	if err != nil {
		return from
	}

	// Start from the next whole minute strictly after `from`, then step one
	// minute at a time until every field matches. The previous approach jumped
	// "60 - minute" minutes on a minute miss, which overshot to the top of the
	// next hour instead of the next matching minute (e.g. "*/5" from 10:02
	// wrongly landed on 11:00 rather than 10:05). Minute stepping is simple and
	// correct; the bound stops an unsatisfiable expression looping forever.
	next := from.Truncate(time.Minute).Add(time.Minute)

	const maxSteps = 366 * 24 * 60 // just over a year of minutes
	for i := 0; i < maxSteps; i++ {
		if matches(next.Minute(), minute) &&
			matches(next.Hour(), hour) &&
			matches(next.Day(), dayOfMonth) &&
			matches(int(next.Month()), month) &&
			matches(int(next.Weekday()), dayOfWeek) {
			return next
		}
		next = next.Add(time.Minute)
	}

	return next
}

func parseField(field string, min, max int) ([]int, error) {
	if field == "*" {
		return allValues(min, max), nil
	}

	var values []int

	for _, part := range strings.Split(field, ",") {
		if strings.Contains(part, "/") {
			step := strings.Split(part, "/")
			if len(step) != 2 {
				continue
			}

			rangePart := step[0]
			stepNum, err := strconv.Atoi(step[1])
			if err != nil {
				continue
			}

			var start, end int
			if rangePart == "*" {
				start = min
				end = max
			} else if strings.Contains(rangePart, "-") {
				rangeParts := strings.Split(rangePart, "-")
				start, _ = strconv.Atoi(rangeParts[0])
				end, _ = strconv.Atoi(rangeParts[1])
			} else {
				start, _ = strconv.Atoi(rangePart)
				end = max
			}

			for i := start; i <= end; i += stepNum {
				values = append(values, i)
			}
		} else if strings.Contains(part, "-") {
			rangeParts := strings.Split(part, "-")
			start, _ := strconv.Atoi(rangeParts[0])
			end, _ := strconv.Atoi(rangeParts[1])
			for i := start; i <= end; i++ {
				values = append(values, i)
			}
		} else {
			val, _ := strconv.Atoi(part)
			values = append(values, val)
		}
	}

	return values, nil
}

func allValues(min, max int) []int {
	result := make([]int, max-min+1)
	for i := min; i <= max; i++ {
		result[i-min] = i
	}
	return result
}

func matches(value int, allowed []int) bool {
	if allowed == nil {
		return true
	}
	for _, a := range allowed {
		if a == value {
			return true
		}
	}
	return false
}

func HumanReadable(expr string) string {
	parts := strings.Fields(expr)
	if len(parts) < 5 {
		return expr
	}

	minute := parts[0]
	hour := parts[1]
	dayOfMonth := parts[2]
	month := parts[3]
	dayOfWeek := parts[4]

	var description []string

	if minute == "*" && hour == "*" {
		description = append(description, "every minute")
	} else if minute == "0" && hour == "*" {
		description = append(description, "every hour")
	} else if minute != "*" && hour != "*" {
		description = append(description, fmt.Sprintf("at %s:%s", hour, minute))
	}

	if dayOfMonth != "*" {
		description = append(description, fmt.Sprintf("on day %s", dayOfMonth))
	}

	if month != "*" {
		monthNames := []string{"", "Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"}
		m, _ := strconv.Atoi(month)
		if m >= 1 && m <= 12 {
			description = append(description, fmt.Sprintf("in %s", monthNames[m]))
		}
	}

	if dayOfWeek != "*" {
		dayNames := []string{"Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"}
		d, _ := strconv.Atoi(dayOfWeek)
		if d >= 0 && d <= 6 {
			description = append(description, fmt.Sprintf("on %s", dayNames[d]))
		}
	}

	if len(description) == 0 {
		return expr
	}

	return strings.Join(description, " ")
}
