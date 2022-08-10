package duffel

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/jace-ys/simple-api/domain"
	"github.com/jace-ys/simple-api/httpapi"
)

var _ domain.FlightsService = (*AirlineBClient)(nil)

type AirlineBClient struct {
	BaseURL *url.URL
	client  *http.Client
}

func NewAirlineBClient() *AirlineBClient {
	url, err := url.Parse("http://interview.duffel.com/airline_b")
	if err != nil {
		panic(err)
	}

	return &AirlineBClient{
		BaseURL: url,
		client:  http.DefaultClient,
	}
}

type flightB struct {
	Arrival      time.Time `json:"arrival"`
	Currency     string    `json:"currency"`
	Departure    time.Time `json:"departure"`
	Dest         string    `json:"dest"`
	FlightNumber string    `json:"flight_number"`
	ID           string    `json:"id"`
	Origin       string    `json:"origin"`
	Price        struct {
		Amount float64 `json:"amount"`
	} `json:"price"`
}

func (f *flightB) toDomain() *domain.DuffelFlight {
	return &domain.DuffelFlight{
		ArrivalTime:     f.Arrival,
		DepartureTime:   f.Departure,
		DurationMinutes: int(f.Arrival.Sub(f.Departure).Minutes()),
		TotalAmount:     f.Price.Amount,
		Currency:        f.Currency,
		FlightNumber:    f.FlightNumber,
		Origin:          f.Origin,
		Destination:     f.Dest,
	}
}

func (c *AirlineBClient) GetFlights(ctx context.Context, origin, destination, departureDate string) (domain.DuffelFlights, error) {
	type payload struct {
		Origin        string `json:"origin"`
		Destination   string `json:"destination"`
		DepartureDate string `json:"departure_date"`
	}

	endpoint := "/"
	req, err := httpapi.NewRequest(ctx, c.BaseURL, http.MethodPost, endpoint, &payload{
		Origin:        origin,
		Destination:   destination,
		DepartureDate: departureDate,
	})

	var res struct {
		Flights []flightB `json:"flights"`
	}
	rsp, err := httpapi.Do(c.client, req, &res)
	if err != nil {
		return nil, err
	}

	switch {
	case rsp.StatusCode == http.StatusOK:
		// OK
	case 500 <= rsp.StatusCode && rsp.StatusCode <= 599:
		return nil, fmt.Errorf("%w: %s", httpapi.ErrDownstreamUnavailable, rsp.HTTPErrorBody)
	default:
		return nil, fmt.Errorf("%w: %d", httpapi.ErrStatusCodeUnknown, rsp.StatusCode)
	}

	flights := make(domain.DuffelFlights, len(res.Flights))
	for i, f := range res.Flights {
		flights[i] = f.toDomain()
	}

	return flights, nil
}
