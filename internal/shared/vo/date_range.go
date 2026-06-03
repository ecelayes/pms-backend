package vo

import (
	"errors"
	"time"
)

var (
	ErrinvalidDateRange = errors.New("start date must be before or equal to end date")
)

type DateRange struct {
	start time.Time
	end   time.Time
}

func NewDateRange(start, end time.Time) (DateRange, error) {
	if start.After(end) {
		return DateRange{}, ErrinvalidDateRange
	}
	return DateRange{start: start, end: end}, nil
}
func (d DateRange) Start() time.Time {
	return d.start
}
func (d DateRange) End() time.Time {
	return d.end
}
func (d DateRange) Overlaps(other DateRange) bool {
	return d.start.Before(other.end) && other.start.Before(d.end)
}
func (d DateRange) DaysCount() int {
	diff := d.end.Sub(d.start)
	return int(diff.Hours() / 24)
}
