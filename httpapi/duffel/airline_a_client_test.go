package duffel_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jace-ys/simple-api/domain"
	"github.com/jace-ys/simple-api/httpapi"
	"github.com/jace-ys/simple-api/httpapi/duffel"
)

func TestAirlineAGetFlights(t *testing.T) {
	tt := []struct {
		Name             string
		DownstreamStatus int
		ExpectedCount    int
		ExpectedError    error
	}{
		{
			Name:             "Returns movies on response status 200",
			DownstreamStatus: http.StatusOK,
			ExpectedCount:    5,
		},
		{
			Name:             "Returns ErrDownstreamUnavailable on response status 500",
			DownstreamStatus: http.StatusInternalServerError,
			ExpectedError:    httpapi.ErrDownstreamUnavailable,
		},
		{
			Name:             "Returns ErrStatusCodeUnknown on unrecognised response status",
			DownstreamStatus: http.StatusUnauthorized,
			ExpectedError:    httpapi.ErrStatusCodeUnknown,
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			handler, client := setupAirlineA(t)

			fixture, err := os.ReadFile("fixtures/airline-a.json")
			assert.NoError(t, err)

			handler.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
				assert.Equal(t, "application/json", r.Header.Get("Accept"))

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tc.DownstreamStatus)
				if tc.DownstreamStatus == 200 {
					w.Write(fixture)
				}
			})

			flights, err := client.GetFlights(context.Background(), "LHR", "JFK", "2019-10-21")

			if tc.ExpectedError != nil {
				assert.ErrorIs(t, err, tc.ExpectedError)
				assert.Nil(t, flights)
			} else {
				assert.NoError(t, err)
				assert.Len(t, flights, tc.ExpectedCount)
			}
		})
	}
}

func setupAirlineA(t *testing.T) (*http.ServeMux, domain.FlightsService) {
	handler := http.NewServeMux()

	server := httptest.NewServer(handler)
	serverURL, err := url.Parse(server.URL)
	assert.NoError(t, err)

	client := duffel.NewAirlineAClient()
	client.BaseURL = serverURL

	return handler, client
}
