package mcu

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/jace-ys/simple-api/domain"
	"github.com/jace-ys/simple-api/httpapi"
)

var _ domain.MoviesService = (*Client)(nil)

type Client struct {
	BaseURL *url.URL
	client  *http.Client
}

func NewClient() *Client {
	url, err := url.Parse("https://mcuapi.herokuapp.com/api/v1/")
	if err != nil {
		panic(err)
	}

	return &Client{
		BaseURL: url,
		client:  http.DefaultClient,
	}
}

type movie struct {
	ID               int    `json:"id"`
	Title            string `json:"title"`
	ReleaseDate      string `json:"release_date"`
	BoxOffice        string `json:"box_office"`
	Duration         int    `json:"duration"`
	Overview         string `json:"overview"`
	CoverURL         string `json:"cover_url"`
	TrailerURL       string `json:"trailer_url"`
	DirectedBy       string `json:"directed_by"`
	Phase            int    `json:"phase"`
	Saga             string `json:"saga"`
	Chronology       int    `json:"chronology"`
	PostCreditScenes int    `json:"post_credit_scenes"`
	ImdbID           string `json:"imdb_id"`
}

func (m *movie) toDomain() (*domain.Movie, error) {
	bo, err := strconv.Atoi(m.BoxOffice)
	if err != nil {
		return nil, err
	}

	return &domain.Movie{
		ID:               m.ID,
		Title:            m.Title,
		ReleaseDate:      m.ReleaseDate,
		BoxOffice:        bo,
		DurationMinutes:  m.Duration,
		Overview:         m.Overview,
		Phase:            m.Phase,
		Saga:             m.Saga,
		Chronology:       m.Chronology,
		PostCreditScenes: m.PostCreditScenes,
	}, nil
}

func (c *Client) GetMovies(ctx context.Context) (domain.Movies, error) {
	endpoint := "/movies"
	req, err := httpapi.NewRequest(ctx, c.BaseURL, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	var res struct {
		Data []movie `json:"data"`
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

	movies := make(domain.Movies, len(res.Data))
	for i, m := range res.Data {
		movie, err := m.toDomain()
		if err != nil {
			return nil, err
		}
		movies[i] = movie
	}

	return movies, nil
}

func (c *Client) GetMovie(ctx context.Context, movieID int) (*domain.Movie, error) {
	endpoint := fmt.Sprintf("/movies/%d", movieID)
	req, err := httpapi.NewRequest(ctx, c.BaseURL, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	var res movie
	rsp, err := httpapi.Do(c.client, req, &res)
	if err != nil {
		return nil, err
	}

	switch {
	case rsp.StatusCode == http.StatusOK:
		// OK
	case rsp.StatusCode == http.StatusNotFound:
		return nil, domain.ErrMovieNotFound
	case 500 <= rsp.StatusCode && rsp.StatusCode <= 599:
		return nil, fmt.Errorf("%w: %s", httpapi.ErrDownstreamUnavailable, rsp.HTTPErrorBody)
	default:
		return nil, fmt.Errorf("%w: %d", httpapi.ErrStatusCodeUnknown, rsp.StatusCode)
	}

	movie, err := res.toDomain()
	if err != nil {
		return nil, err
	}

	return movie, nil
}
