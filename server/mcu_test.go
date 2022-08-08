package server_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"

	"github.com/jace-ys/simple-api/domain"
	"github.com/jace-ys/simple-api/domain/domainfakes"
	"github.com/jace-ys/simple-api/server"
)

func init() {
	log.SetOutput(io.Discard)
}

func TestGetMovies(t *testing.T) {
	movie := &domain.Movie{
		ID:               4,
		Title:            "Hello World",
		ReleaseDate:      "2000-01-01",
		BoxOffice:        1000000,
		DurationMinutes:  120,
		Overview:         "This is a fantastic movie.",
		Phase:            1,
		Saga:             "Finale",
		Chronology:       4,
		PostCreditScenes: 2,
	}

	tt := []struct {
		Name           string
		SetupFake      func(fake *domainfakes.FakeMoviesService)
		ExpectedStatus int
		ExpectedBody   []*domain.Movie
	}{
		{
			Name: "Returns status 200",
			SetupFake: func(fake *domainfakes.FakeMoviesService) {
				fake.GetMoviesReturns([]*domain.Movie{movie}, nil)
			},
			ExpectedStatus: http.StatusOK,
			ExpectedBody:   []*domain.Movie{movie},
		},
		{
			Name: "Returns status 500 when service request fails",
			SetupFake: func(fake *domainfakes.FakeMoviesService) {
				fake.GetMoviesReturns(nil, errors.New("internal server error"))
			},
			ExpectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			service := new(domainfakes.FakeMoviesService)
			tc.SetupFake(service)

			router := mux.NewRouter()
			handler := server.NewMCUHandler(service)
			handler.RegisterRoutes(router)

			req, err := http.NewRequest("GET", "/movies", nil)
			assert.NoError(t, err)

			rw := httptest.NewRecorder()
			router.ServeHTTP(rw, req)

			assert.Equal(t, tc.ExpectedStatus, rw.Code)

			if tc.ExpectedBody != nil {
				var res []*domain.Movie
				json.NewDecoder(rw.Body).Decode(&res)
				assert.ElementsMatch(t, tc.ExpectedBody, res)
			}
		})
	}
}

func TestGetMovie(t *testing.T) {
	movie := &domain.Movie{
		ID:               4,
		Title:            "Hello World",
		ReleaseDate:      "2000-01-01",
		BoxOffice:        1000000,
		DurationMinutes:  120,
		Overview:         "This is a fantastic movie.",
		Phase:            1,
		Saga:             "Finale",
		Chronology:       4,
		PostCreditScenes: 2,
	}

	tt := []struct {
		Name           string
		PathParamID    string
		SetupFake      func(fake *domainfakes.FakeMoviesService)
		ExpectedStatus int
		ExpectedBody   *domain.Movie
	}{
		{
			Name:        "Returns status 200",
			PathParamID: "4",
			SetupFake: func(fake *domainfakes.FakeMoviesService) {
				fake.GetMovieReturns(movie, nil)
			},
			ExpectedStatus: http.StatusOK,
			ExpectedBody:   movie,
		},
		{
			Name:           "Returns status 400 when ID is invalid",
			PathParamID:    "test",
			ExpectedStatus: http.StatusBadRequest,
		},
		{
			Name:        "Returns status 404 when not found",
			PathParamID: "4",
			SetupFake: func(fake *domainfakes.FakeMoviesService) {
				fake.GetMovieReturns(nil, domain.ErrMovieNotFound)
			},
			ExpectedStatus: http.StatusNotFound,
		},
		{
			Name:        "Returns status 500 when service request fails",
			PathParamID: "4",
			SetupFake: func(fake *domainfakes.FakeMoviesService) {
				fake.GetMovieReturns(nil, errors.New("internal server error"))
			},
			ExpectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			service := new(domainfakes.FakeMoviesService)
			if tc.SetupFake != nil {
				tc.SetupFake(service)
			}

			router := mux.NewRouter()
			handler := server.NewMCUHandler(service)
			handler.RegisterRoutes(router)

			endpoint := fmt.Sprintf("/movies/%s", tc.PathParamID)
			req, err := http.NewRequest("GET", endpoint, nil)
			assert.NoError(t, err)

			rw := httptest.NewRecorder()
			router.ServeHTTP(rw, req)

			assert.Equal(t, tc.ExpectedStatus, rw.Code)

			if tc.ExpectedBody != nil {
				var res *domain.Movie
				json.NewDecoder(rw.Body).Decode(&res)
				assert.Equal(t, tc.ExpectedBody, res)
			}
		})
	}
}
