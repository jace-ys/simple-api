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
		ExpectedBody   domain.Movies
	}{
		{
			Name: "Returns status 200",
			SetupFake: func(fake *domainfakes.FakeMoviesService) {
				fake.GetMoviesReturns(domain.Movies{movie}, nil)
			},
			ExpectedStatus: http.StatusOK,
			ExpectedBody:   domain.Movies{movie},
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
				var res domain.Movies
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

func TestGetSagas(t *testing.T) {
	movie1 := &domain.Movie{
		ID:               4,
		Title:            "Hello World 4",
		ReleaseDate:      "2000-01-01",
		BoxOffice:        1000000,
		DurationMinutes:  120,
		Overview:         "This is a fantastic movie.",
		Phase:            1,
		Saga:             "Epilogue",
		Chronology:       4,
		PostCreditScenes: 2,
	}
	movie2 := &domain.Movie{
		ID:               6,
		Title:            "Hello World 6",
		ReleaseDate:      "2010-01-01",
		BoxOffice:        1000000,
		DurationMinutes:  120,
		Overview:         "This is another fantastic movie.",
		Phase:            2,
		Saga:             "Epilogue",
		Chronology:       6,
		PostCreditScenes: 2,
	}
	movie3 := &domain.Movie{
		ID:               10,
		Title:            "Hello World 10",
		ReleaseDate:      "2020-01-01",
		BoxOffice:        1000000,
		DurationMinutes:  120,
		Overview:         "This is another another fantastic movie.",
		Phase:            4,
		Saga:             "Finale",
		Chronology:       10,
		PostCreditScenes: 2,
	}

	tt := []struct {
		Name           string
		QueryParam     string
		SetupFake      func(fake *domainfakes.FakeMoviesService)
		ExpectedStatus int
		ExpectedBody   interface{}
	}{
		{
			Name: "Returns status 200 without query param",
			SetupFake: func(fake *domainfakes.FakeMoviesService) {
				fake.GetMoviesReturns(domain.Movies{movie1, movie2, movie3}, nil)
			},
			ExpectedStatus: http.StatusOK,
			ExpectedBody: domain.Sagas{
				{
					Name:                 "Epilogue",
					StartDate:            "2000-01-01",
					EndDate:              "2010-01-01",
					TotalBoxOffice:       2000000,
					TotalDurationMinutes: 240,
					TotalMovies:          2,
					Phases: domain.Phases{
						{
							Number: 1,
							Movies: domain.Movies{movie1},
						},
						{
							Number: 2,
							Movies: domain.Movies{movie2},
						},
					},
				},
				{
					Name:                 "Finale",
					StartDate:            "2020-01-01",
					EndDate:              "2020-01-01",
					TotalBoxOffice:       1000000,
					TotalDurationMinutes: 120,
					TotalMovies:          1,
					Phases: domain.Phases{
						{
							Number: 4,
							Movies: domain.Movies{movie3},
						},
					},
				},
			},
		},
		{
			Name:       "Returns status 200 with query param",
			QueryParam: "?name=epilogue",
			SetupFake: func(fake *domainfakes.FakeMoviesService) {
				fake.GetMoviesReturns(domain.Movies{movie1, movie2, movie3}, nil)
			},
			ExpectedStatus: http.StatusOK,
			ExpectedBody: &domain.Saga{
				Name:                 "Epilogue",
				StartDate:            "2000-01-01",
				EndDate:              "2010-01-01",
				TotalBoxOffice:       2000000,
				TotalDurationMinutes: 240,
				TotalMovies:          2,
				Phases: domain.Phases{
					{
						Number: 1,
						Movies: domain.Movies{movie1},
					},
					{
						Number: 2,
						Movies: domain.Movies{movie2},
					},
				},
			},
		},
		{
			Name:           "Returns status 404 when not found",
			QueryParam:     "?name=invalid",
			ExpectedStatus: http.StatusNotFound,
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
			if tc.SetupFake != nil {
				tc.SetupFake(service)
			}

			router := mux.NewRouter()
			handler := server.NewMCUHandler(service)
			handler.RegisterRoutes(router)

			endpoint := fmt.Sprintf("/sagas%s", tc.QueryParam)
			req, err := http.NewRequest("GET", endpoint, nil)
			assert.NoError(t, err)

			rw := httptest.NewRecorder()
			router.ServeHTTP(rw, req)

			assert.Equal(t, tc.ExpectedStatus, rw.Code)

			if tc.ExpectedBody != nil {
				if tc.QueryParam == "" {
					var res domain.Sagas
					json.NewDecoder(rw.Body).Decode(&res)
					assert.ElementsMatch(t, tc.ExpectedBody, res)
				} else {
					var res *domain.Saga
					json.NewDecoder(rw.Body).Decode(&res)
					assert.Equal(t, tc.ExpectedBody, res)
				}
			}
		})
	}
}
