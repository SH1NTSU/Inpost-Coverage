package domain

import "time"

type AvailabilitySnapshot struct {
	ID         int64
	PointID    int64
	CapturedAt time.Time
	Status     PointStatus
}
