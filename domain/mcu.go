package domain

import (
	"context"
	"errors"
	"sort"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var (
	ErrSagaNotFound  = errors.New("saga not found")
	ErrMovieNotFound = errors.New("movie not found")
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . MoviesService
type MoviesService interface {
	GetMovies(ctx context.Context) (Movies, error)
	GetMovie(ctx context.Context, movieID int) (*Movie, error)
}

type Sagas []*Saga

func (s Sagas) Len() int           { return len(s) }
func (s Sagas) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s Sagas) Less(i, j int) bool { return s[i].StartDate < s[j].StartDate }

type Saga struct {
	Name                 string `json:"name"`
	StartDate            string `json:"start_date"`
	EndDate              string `json:"end_date"`
	TotalBoxOffice       int    `json:"total_box_office"`
	TotalDurationMinutes int    `json:"total_duration_minutes"`
	TotalMovies          int    `json:"total_movies"`
	Phases               Phases `json:"phases"`
}

type Phases []*Phase

func (s Phases) Len() int           { return len(s) }
func (s Phases) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s Phases) Less(i, j int) bool { return s[i].Number < s[j].Number }

type Phase struct {
	Number int    `json:"number"`
	Movies Movies `json:"movies"`
}

type Movies []*Movie

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

func (m Movies) GroupBySaga() []*Saga {
	sm := make(map[string]Movies)
	for _, movie := range m {
		sm[movie.Saga] = append(sm[movie.Saga], movie)
	}

	sagas := Sagas{}
	for name, sagaMovies := range sm {
		var boxOffice int
		var durationMinutes int

		for _, movie := range sagaMovies {
			boxOffice += movie.BoxOffice
			durationMinutes += movie.DurationMinutes
		}

		sagas = append(sagas, &Saga{
			Name:                 name,
			StartDate:            sagaMovies[0].ReleaseDate,
			EndDate:              sagaMovies[len(sagaMovies)-1].ReleaseDate,
			TotalBoxOffice:       boxOffice,
			TotalDurationMinutes: durationMinutes,
			TotalMovies:          len(sagaMovies),
			Phases:               sagaMovies.GroupByPhase(),
		})
	}

	sort.Sort(sagas)
	return sagas
}

func (m Movies) GetSaga(name string) (*Saga, error) {
	sm := make(map[string]Movies)
	for _, movie := range m {
		sm[movie.Saga] = append(sm[movie.Saga], movie)
	}

	sagaName := cases.Title(language.English).String(name)
	sagaMovies, ok := sm[sagaName]
	if !ok {
		return nil, ErrSagaNotFound
	}

	saga := &Saga{
		Name:        sagaName,
		StartDate:   sagaMovies[0].ReleaseDate,
		EndDate:     sagaMovies[len(sagaMovies)-1].ReleaseDate,
		TotalMovies: len(sagaMovies),
		Phases:      sagaMovies.GroupByPhase(),
	}

	for _, movie := range sagaMovies {
		saga.TotalBoxOffice += movie.BoxOffice
		saga.TotalDurationMinutes += movie.DurationMinutes
	}

	return saga, nil
}

func (m Movies) GroupByPhase() Phases {
	pm := make(map[int]Movies)
	for _, movie := range m {
		pm[movie.Phase] = append(pm[movie.Phase], movie)
	}

	phases := Phases{}
	for num, phaseMovies := range pm {
		phases = append(phases, &Phase{
			Number: num,
			Movies: phaseMovies,
		})
	}

	sort.Sort(phases)
	return phases
}
