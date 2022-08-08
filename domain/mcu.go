package domain

import (
	"context"
	"errors"
)

var (
	ErrMovieNotFound = errors.New("movie not found")
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . MoviesService
type MoviesService interface {
	GetMovies(ctx context.Context) ([]*Movie, error)
	GetMovie(ctx context.Context, movieID int) (*Movie, error)
}

type Movie struct {
	ID               int    `json:"id"`
	Title            string `json:"title"`
	ReleaseDate      string `json:"release_date"`
	BoxOffice        int    `json:"box_office"`
	DurationMinutes  int    `json:"duration_minutse"`
	Overview         string `json:"overview"`
	Phase            int    `json:"phase"`
	Saga             string `json:"saga"`
	Chronology       int    `json:"chronology"`
	PostCreditScenes int    `json:"post_credit_scenes"`
}
