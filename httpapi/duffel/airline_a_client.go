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

var _ domain.FlightsService = (*AirlineAClient)(nil)

type AirlineAClient struct {
	BaseURL *url.URL
	client  *http.Client
}

func NewAirlineAClient() *AirlineAClient {
	url, err := url.Parse("http://interview.duffel.com/airline_a")
	if err != nil {
		panic(err)
	}

	return &AirlineAClient{
		BaseURL: url,
		client:  http.DefaultClient,
	}
}

type flightA struct {
	Arrival       time.Time `json:"arrival"`
	Departure     time.Time `json:"departure"`
	Destination   string    `json:"destination"`
	Duration      int       `json:"duration"`
	FlightNumber  string    `json:"flight_number"`
	ID            string    `json:"id"`
	Origin        string    `json:"origin"`
	TotalAmount   float64   `json:"total_amount"`
	TotalCurrency string    `json:"total_currency"`
}

func (f *flightA) toDomain() *domain.DuffelFlight {
	return &domain.DuffelFlight{
		ArrivalTime:     f.Arrival,
		DepartureTime:   f.Departure,
		DurationMinutes: f.Duration,
		TotalAmount:     f.TotalAmount / 100.00,
		Currency:        f.TotalCurrency,
		FlightNumber:    f.FlightNumber,
		Origin:          f.Origin,
		Destination:     f.Destination,
	}
}

func (c *AirlineAClient) GetFlights(ctx context.Context, origin, destination, departureDate string) (domain.DuffelFlights, error) {
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
	if err != nil {
		return nil, err
	}

	var res struct {
		Data struct {
			Offers []flightA `json:"offers"`
		} `json:"data"`
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

	flights := make(domain.DuffelFlights, len(res.Data.Offers))
	for i, f := range res.Data.Offers {
		flights[i] = f.toDomain()
	}

	return flights, nil
}
