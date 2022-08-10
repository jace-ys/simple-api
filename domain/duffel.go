package domain

import (
	"context"
	"sort"
	"time"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . FlightsService
type FlightsService interface {
	GetFlights(ctx context.Context, origin, destination, departureDate string) (DuffelFlights, error)
}

type DuffelFlights []*DuffelFlight

type DuffelFlight struct {
	ArrivalTime     time.Time `json:"arrival_time"`
	DepartureTime   time.Time `json:"departure_time"`
	DurationMinutes int       `json:"duration_minutes"`
	TotalAmount     float64   `json:"total_amount"`
	Currency        string    `json:"currency"`
	FlightNumber    string    `json:"flight_number"`
	Origin          string    `json:"origin"`
	Destination     string    `json:"destination"`
}

type SortOrder string

var (
	SortAsc  SortOrder = "asc"
	SortDesc SortOrder = "desc"
)

func (f DuffelFlights) SortByPrice(order SortOrder) DuffelFlights {
	s := sortAscPrice(f)
	switch order {
	case SortAsc:
		sort.Sort(s)
	case SortDesc:
		sort.Sort(sort.Reverse(s))
	}
	return DuffelFlights(s)
}

func (f DuffelFlights) SortByDuration(order SortOrder) DuffelFlights {
	s := sortAscDuration(f)
	switch order {
	case SortAsc:
		sort.Sort(s)
	case SortDesc:
		sort.Sort(sort.Reverse(s))
	}
	return DuffelFlights(s)
}

type sortAscPrice DuffelFlights

func (s sortAscPrice) Len() int           { return len(s) }
func (s sortAscPrice) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s sortAscPrice) Less(i, j int) bool { return s[i].TotalAmount < s[j].TotalAmount }

type sortAscDuration DuffelFlights

func (s sortAscDuration) Len() int           { return len(s) }
func (s sortAscDuration) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s sortAscDuration) Less(i, j int) bool { return s[i].DurationMinutes < s[j].DurationMinutes }
