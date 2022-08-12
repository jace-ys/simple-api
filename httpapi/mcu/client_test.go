package mcu_test

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
	"github.com/jace-ys/simple-api/httpapi/mcu"
)

func TestGetMovies(t *testing.T) {
	movie1 := &domain.Movie{
		ID:               1,
		Title:            "Iron Man",
		ReleaseDate:      "2008-05-02",
		BoxOffice:        585171547,
		DurationMinutes:  126,
		Overview:         "2008's Iron Man tells the story of Tony Stark, a billionaire industrialist and genius inventor who is kidnapped and forced to build a devastating weapon. Instead, using his intelligence and ingenuity, Tony builds a high-tech suit of armor and escapes captivity. When he uncovers a nefarious plot with global implications, he dons his powerful armor and vows to protect the world as Iron Man.",
		Phase:            1,
		Saga:             "Infinity Saga",
		Chronology:       3,
		PostCreditScenes: 1,
	}
	movie2 := &domain.Movie{
		ID:               2,
		Title:            "The Incredible Hulk",
		ReleaseDate:      "2008-06-13",
		BoxOffice:        265573859,
		DurationMinutes:  112,
		Overview:         "In this new beginning, scientist Bruce Banner desperately hunts for a cure to the gamma radiation that poisoned his cells and unleashes the unbridled force of rage within him: The Hulk. Living in the shadows--cut off from a life he knew and the woman he loves, Betty Ross--Banner struggles to avoid the obsessive pursuit of his nemesis, General Thunderbolt Ross and the military machinery that seeks to capture him and brutally exploit his power. As all three grapple with the secrets that led to the Hulk's creation, they are confronted with a monstrous new adversary known as the Abomination, whose destructive strength exceeds even the Hulk's own. One scientist must make an agonizing final choice: accept a peaceful life as Bruce Banner or find heroism in the creature he holds inside--The Incredible Hulk.",
		Phase:            1,
		Saga:             "Infinity Saga",
		Chronology:       5,
		PostCreditScenes: 1,
	}

	tt := []struct {
		Name             string
		DownstreamStatus int
		ExpectedBody     domain.Movies
		ExpectedError    error
	}{
		{
			Name:             "Returns movies on response status 200",
			DownstreamStatus: http.StatusOK,
			ExpectedBody:     domain.Movies{movie1, movie2},
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
			handler, client := setupMCU(t)

			fixture, err := os.ReadFile("fixtures/list-movies.json")
			assert.NoError(t, err)

			handler.HandleFunc("/movies", func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
				assert.Equal(t, "application/json", r.Header.Get("Accept"))

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tc.DownstreamStatus)
				if tc.DownstreamStatus == 200 {
					w.Write(fixture)
				}
			})

			movies, err := client.GetMovies(context.Background())

			if tc.ExpectedError != nil {
				assert.ErrorIs(t, err, tc.ExpectedError)
				assert.Nil(t, movies)
			} else {
				assert.NoError(t, err)
				assert.ElementsMatch(t, tc.ExpectedBody, movies)
			}
		})
	}
}

func TestGetMovie(t *testing.T) {
	movie := &domain.Movie{
		ID:               1,
		Title:            "Iron Man",
		ReleaseDate:      "2008-05-02",
		BoxOffice:        585171547,
		DurationMinutes:  126,
		Overview:         "2008's Iron Man tells the story of Tony Stark, a billionaire industrialist and genius inventor who is kidnapped and forced to build a devastating weapon. Instead, using his intelligence and ingenuity, Tony builds a high-tech suit of armor and escapes captivity. When he uncovers a nefarious plot with global implications, he dons his powerful armor and vows to protect the world as Iron Man.",
		Phase:            1,
		Saga:             "Infinity Saga",
		Chronology:       3,
		PostCreditScenes: 1,
	}

	tt := []struct {
		Name             string
		DownstreamStatus int
		ExpectedBody     *domain.Movie
		ExpectedError    error
	}{
		{
			Name:             "Returns movie on response status 200",
			DownstreamStatus: http.StatusOK,
			ExpectedBody:     movie,
		},
		{
			Name:             "Returns domain.ErrMovieNotFound on response status 404",
			DownstreamStatus: http.StatusNotFound,
			ExpectedError:    domain.ErrMovieNotFound,
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
			handler, client := setupMCU(t)

			fixture, err := os.ReadFile("fixtures/get-movie.json")
			assert.NoError(t, err)

			handler.HandleFunc("/movies/1", func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
				assert.Equal(t, "application/json", r.Header.Get("Accept"))

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tc.DownstreamStatus)
				if tc.DownstreamStatus == 200 {
					w.Write(fixture)
				}
			})

			movie, err := client.GetMovie(context.Background(), 1)

			if tc.ExpectedError != nil {
				assert.ErrorIs(t, err, tc.ExpectedError)
				assert.Nil(t, movie)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.ExpectedBody, movie)
			}
		})
	}
}

func setupMCU(t *testing.T) (*http.ServeMux, domain.MoviesService) {
	handler := http.NewServeMux()

	server := httptest.NewServer(handler)
	serverURL, err := url.Parse(server.URL)
	assert.NoError(t, err)

	client := mcu.NewClient()
	client.BaseURL = serverURL

	return handler, client
}
