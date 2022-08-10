package server_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"

	"github.com/jace-ys/simple-api/domain"
	"github.com/jace-ys/simple-api/domain/domainfakes"
	"github.com/jace-ys/simple-api/server"
)

func init() {
	log.SetOutput(io.Discard)
}

func TestSearchFlights(t *testing.T) {
	ft, err := time.Parse(time.RFC3339, "2006-01-02T15:04:05Z")
	assert.NoError(t, err)

	flightsA := domain.DuffelFlights{
		{
			ArrivalTime:     ft,
			DepartureTime:   ft,
			DurationMinutes: 1,
			TotalAmount:     20.00,
			Currency:        "GBP",
			FlightNumber:    "123",
			Origin:          "LHR",
			Destination:     "JFK",
		},
		{
			ArrivalTime:     ft,
			DepartureTime:   ft,
			DurationMinutes: 3,
			TotalAmount:     10.00,
			Currency:        "GBP",
			FlightNumber:    "123",
			Origin:          "LHR",
			Destination:     "JFK",
		},
	}

	flightsB := domain.DuffelFlights{
		{
			ArrivalTime:     ft,
			DepartureTime:   ft,
			DurationMinutes: 2,
			TotalAmount:     40.00,
			Currency:        "GBP",
			FlightNumber:    "456",
			Origin:          "LHR",
			Destination:     "JFK",
		},
		{
			ArrivalTime:     ft,
			DepartureTime:   ft,
			DurationMinutes: 4,
			TotalAmount:     30.00,
			Currency:        "GBP",
			FlightNumber:    "456",
			Origin:          "LHR",
			Destination:     "JFK",
		},
	}

	tt := []struct {
		Name           string
		QueryParams    string
		SetupFakeA     func(fake *domainfakes.FakeFlightsService)
		SetupFakeB     func(fake *domainfakes.FakeFlightsService)
		ReqBody        *server.SearchFlightsRequest
		ExpectedStatus int
		ExpectedBody   domain.DuffelFlights
	}{
		{
			Name: "Returns status 200",
			SetupFakeA: func(fake *domainfakes.FakeFlightsService) {
				fake.GetFlightsReturns(flightsA, nil)
			},
			SetupFakeB: func(fake *domainfakes.FakeFlightsService) {
				fake.GetFlightsReturns(flightsB, nil)
			},
			ReqBody: &server.SearchFlightsRequest{
				Origin:        "LHR",
				Destination:   "JFK",
				DepartureDate: "2019-10-21",
			},
			ExpectedStatus: http.StatusOK,
			ExpectedBody:   append(flightsA, flightsB...),
		},
		{
			Name:        "Returns status 200 with ascending price",
			QueryParams: "?sort_by=price&order=asc",
			SetupFakeA: func(fake *domainfakes.FakeFlightsService) {
				fake.GetFlightsReturns(flightsA, nil)
			},
			SetupFakeB: func(fake *domainfakes.FakeFlightsService) {
				fake.GetFlightsReturns(flightsB, nil)
			},
			ReqBody: &server.SearchFlightsRequest{
				Origin:        "LHR",
				Destination:   "JFK",
				DepartureDate: "2019-10-21",
			},
			ExpectedStatus: http.StatusOK,
			ExpectedBody:   domain.DuffelFlights{flightsA[1], flightsA[0], flightsB[1], flightsB[0]},
		},
		{
			Name:        "Returns status 200 with descending price",
			QueryParams: "?sort_by=price&order=desc",
			SetupFakeA: func(fake *domainfakes.FakeFlightsService) {
				fake.GetFlightsReturns(flightsA, nil)
			},
			SetupFakeB: func(fake *domainfakes.FakeFlightsService) {
				fake.GetFlightsReturns(flightsB, nil)
			},
			ReqBody: &server.SearchFlightsRequest{
				Origin:        "LHR",
				Destination:   "JFK",
				DepartureDate: "2019-10-21",
			},
			ExpectedStatus: http.StatusOK,
			ExpectedBody:   domain.DuffelFlights{flightsB[0], flightsB[1], flightsA[0], flightsA[1]},
		},
		{
			Name:        "Returns status 200 with ascending duration",
			QueryParams: "?sort_by=duration&order=asc",
			SetupFakeA: func(fake *domainfakes.FakeFlightsService) {
				fake.GetFlightsReturns(flightsA, nil)
			},
			SetupFakeB: func(fake *domainfakes.FakeFlightsService) {
				fake.GetFlightsReturns(flightsB, nil)
			},
			ReqBody: &server.SearchFlightsRequest{
				Origin:        "LHR",
				Destination:   "JFK",
				DepartureDate: "2019-10-21",
			},
			ExpectedStatus: http.StatusOK,
			ExpectedBody:   domain.DuffelFlights{flightsA[0], flightsB[0], flightsA[1], flightsB[1]},
		},
		{
			Name:        "Returns status 200 with descending duration",
			QueryParams: "?sort_by=duration&order=desc",
			SetupFakeA: func(fake *domainfakes.FakeFlightsService) {
				fake.GetFlightsReturns(flightsA, nil)
			},
			SetupFakeB: func(fake *domainfakes.FakeFlightsService) {
				fake.GetFlightsReturns(flightsB, nil)
			},
			ReqBody: &server.SearchFlightsRequest{
				Origin:        "LHR",
				Destination:   "JFK",
				DepartureDate: "2019-10-21",
			},
			ExpectedStatus: http.StatusOK,
			ExpectedBody:   domain.DuffelFlights{flightsB[1], flightsA[1], flightsB[0], flightsA[0]},
		},
		{
			Name: "Returns status 200 with partial response when one service request fails",
			SetupFakeA: func(fake *domainfakes.FakeFlightsService) {
				fake.GetFlightsReturns(flightsA, nil)
			},
			SetupFakeB: func(fake *domainfakes.FakeFlightsService) {
				fake.GetFlightsReturns(nil, errors.New("internal server error"))
			},
			ReqBody: &server.SearchFlightsRequest{
				Origin:        "LHR",
				Destination:   "JFK",
				DepartureDate: "2019-10-21",
			},
			ExpectedStatus: http.StatusOK,
			ExpectedBody:   flightsA,
		},
		{
			Name: "Returns status 400 when origin is invalid",
			ReqBody: &server.SearchFlightsRequest{
				Origin:        "invalid",
				Destination:   "JFK",
				DepartureDate: "2019-10-21",
			},
			ExpectedStatus: http.StatusBadRequest,
		},
		{
			Name: "Returns status 400 when origin is invalid",
			ReqBody: &server.SearchFlightsRequest{
				Origin:        "LHR",
				Destination:   "invalid",
				DepartureDate: "2019-10-21",
			},
			ExpectedStatus: http.StatusBadRequest,
		},
		{
			Name: "Returns status 400 when departure date is invalid",
			ReqBody: &server.SearchFlightsRequest{
				Origin:        "LHR",
				Destination:   "invalid",
				DepartureDate: "2019",
			},
			ExpectedStatus: http.StatusBadRequest,
		},
		{
			Name: "Returns status 500 when both service requests fail",
			SetupFakeA: func(fake *domainfakes.FakeFlightsService) {
				fake.GetFlightsReturns(nil, errors.New("internal server error"))
			},
			SetupFakeB: func(fake *domainfakes.FakeFlightsService) {
				fake.GetFlightsReturns(nil, errors.New("internal server error"))
			},
			ReqBody: &server.SearchFlightsRequest{
				Origin:        "LHR",
				Destination:   "JFK",
				DepartureDate: "2019-10-21",
			},
			ExpectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			serviceA := new(domainfakes.FakeFlightsService)
			if tc.SetupFakeA != nil {
				tc.SetupFakeA(serviceA)
			}

			serviceB := new(domainfakes.FakeFlightsService)
			if tc.SetupFakeB != nil {
				tc.SetupFakeB(serviceB)
			}

			router := mux.NewRouter()
			handler := server.NewDuffelFlightsHandler(serviceA, serviceB)
			handler.RegisterRoutes(router)

			body, err := json.Marshal(tc.ReqBody)
			assert.NoError(t, err)

			req, err := http.NewRequest("POST", "/flights/search", bytes.NewBuffer(body))
			assert.NoError(t, err)

			rw := httptest.NewRecorder()
			router.ServeHTTP(rw, req)

			assert.Equal(t, tc.ExpectedStatus, rw.Code)

			if tc.ExpectedBody != nil {
				var res domain.DuffelFlights
				json.NewDecoder(rw.Body).Decode(&res)
				assert.ElementsMatch(t, tc.ExpectedBody, res)
			}
		})
	}
}
