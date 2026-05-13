package domain

import "time"

type PointStatus string

const (
	StatusOperating PointStatus = "Operating"
	StatusDisabled  PointStatus = "Disabled"
)

type Point struct {
	ID           int64
	InpostID     string
	Country      string
	Status       PointStatus
	Latitude     float64
	Longitude    float64
	City         string
	Province     string
	PostCode     string
	Street       string
	BuildingNo   string
	LocationType string
	IsNext       bool
	Location247  bool
	PhysicalType string
	ImageURL     string
	UpdatedAt    time.Time
}
