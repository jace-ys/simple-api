package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/jace-ys/simple-api/domain"
)

var _ domain.MoviesService = (*MCUClient)(nil)

var (
	ErrDownstreamUnavailable = errors.New("downstream unavailable")
	ErrStatusCodeUnknown     = errors.New("unexpected response code")
)

type MCUClient struct {
	BaseURL *url.URL
	client  *http.Client
}

func NewMCUClient() *MCUClient {
	url, err := url.Parse("https://mcuapi.herokuapp.com/api/v1/")
	if err != nil {
		panic(err)
	}

	return &MCUClient{
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

func (c *MCUClient) GetMovies(ctx context.Context) ([]*domain.Movie, error) {
	endpoint := "/movies"
	req, err := c.newRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	var res struct {
		Data []movie `json:"data"`
	}
	rsp, err := c.do(req, &res)
	if err != nil {
		return nil, err
	}

	switch {
	case rsp.StatusCode == http.StatusOK:
		// OK
	case 500 <= rsp.StatusCode && rsp.StatusCode <= 599:
		return nil, fmt.Errorf("%w: %s", ErrDownstreamUnavailable, rsp.HTTPErrorBody)
	default:
		return nil, fmt.Errorf("%w: %d", ErrStatusCodeUnknown, rsp.StatusCode)
	}

	movies := make([]*domain.Movie, len(res.Data))
	for i, m := range res.Data {
		movie, err := m.toDomain()
		if err != nil {
			return nil, err
		}
		movies[i] = movie
	}

	return movies, nil
}

func (c *MCUClient) GetMovie(ctx context.Context, movieID int) (*domain.Movie, error) {
	endpoint := fmt.Sprintf("/movies/%d", movieID)
	req, err := c.newRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	var res movie
	rsp, err := c.do(req, &res)
	if err != nil {
		return nil, err
	}

	switch {
	case rsp.StatusCode == http.StatusOK:
		// OK
	case rsp.StatusCode == http.StatusNotFound:
		return nil, domain.ErrMovieNotFound
	case 500 <= rsp.StatusCode && rsp.StatusCode <= 599:
		return nil, fmt.Errorf("%w: %s", ErrDownstreamUnavailable, rsp.HTTPErrorBody)
	default:
		return nil, fmt.Errorf("%w: %d", ErrStatusCodeUnknown, rsp.StatusCode)
	}

	movie, err := res.toDomain()
	if err != nil {
		return nil, err
	}

	return movie, nil
}

func (c *MCUClient) newRequest(ctx context.Context, method, endpoint string, body interface{}) (*http.Request, error) {
	requestURL, err := c.BaseURL.Parse(strings.Trim(endpoint, "/"))
	if err != nil {
		return nil, err
	}

	var buf io.ReadWriter
	if body != nil {
		buf = &bytes.Buffer{}
		err := json.NewEncoder(buf).Encode(body)
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, requestURL.String(), buf)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	return req, nil
}

type Response struct {
	*http.Response
	HTTPErrorBody string
}

func (c *MCUClient) do(req *http.Request, v interface{}) (*Response, error) {
	rsp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()

	if rsp.StatusCode < 200 || rsp.StatusCode > 299 {
		body, err := ioutil.ReadAll(rsp.Body)
		if err != nil {
			return nil, err
		}

		return &Response{
			Response:      rsp,
			HTTPErrorBody: string(body),
		}, nil
	}

	if v != nil {
		err = json.NewDecoder(rsp.Body).Decode(v)
		if err != nil && !errors.Is(err, io.EOF) {
			return nil, err
		}
	}

	return &Response{
		Response: rsp,
	}, nil
}
