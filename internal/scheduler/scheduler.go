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

	next := from.Truncate(time.Minute)

	for {
		if !matches(next.Minute(), minute) {
			next = next.Add(time.Duration(60-next.Minute()) * time.Minute)
			continue
		}

		if !matches(next.Hour(), hour) {
			next = next.Add(time.Duration(60-next.Hour()) * time.Hour)
			next = time.Date(next.Year(), next.Month(), next.Day(), next.Hour(), 0, 0, 0, next.Location())
			continue
		}

		if !matches(int(next.Day()), dayOfMonth) {
			next = next.Add(24 * time.Hour)
			next = time.Date(next.Year(), next.Month(), next.Day(), next.Hour(), 0, 0, 0, next.Location())
			continue
		}

		if !matches(int(next.Month()), month) {
			next = next.AddDate(0, 1, 0)
			next = time.Date(next.Year(), next.Month(), 1, next.Hour(), 0, 0, 0, next.Location())
			continue
		}

		if !matches(int(next.Weekday()), dayOfWeek) {
			next = next.Add(24 * time.Hour)
			continue
		}

		break
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
