package domain

import (
	"fmt"
	"regexp"
	"time"
)

const monthLayout = "01-2006"

var monthPattern = regexp.MustCompile(`^(0[1-9]|1[0-2])-\d{4}$`)

type Month struct {
	value time.Time
}

func ParseMonth(input string) (Month, error) {
	if !monthPattern.MatchString(input) {
		return Month{}, fmt.Errorf("month must have MM-YYYY format")
	}

	value, err := time.Parse(monthLayout, input)
	if err != nil {
		return Month{}, fmt.Errorf("parse month: %w", err)
	}

	return Month{value: value.UTC()}, nil
}

func NewMonth(value time.Time) Month {
	year, month, _ := value.Date()
	return Month{value: time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)}
}

func (m Month) Time() time.Time {
	return m.value
}

func (m Month) String() string {
	return m.value.Format(monthLayout)
}

func (m Month) Before(other Month) bool {
	return m.value.Before(other.value)
}

func (m Month) After(other Month) bool {
	return m.value.After(other.value)
}

func (m Month) MonthsUntilInclusive(other Month) (int, error) {
	if m.After(other) {
		return 0, fmt.Errorf("start month must not be after end month")
	}

	startYear, startMonth, _ := m.value.Date()
	endYear, endMonth, _ := other.value.Date()

	months := (int(endYear)-int(startYear))*12 + int(endMonth) - int(startMonth) + 1
	return months, nil
}
