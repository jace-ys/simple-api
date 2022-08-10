package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

var (
	ErrDownstreamUnavailable = errors.New("downstream unavailable")
	ErrStatusCodeUnknown     = errors.New("unexpected response code")
)

func NewRequest(ctx context.Context, baseURL *url.URL, method, endpoint string, body interface{}) (*http.Request, error) {
	requestURL, err := baseURL.Parse(strings.Trim(endpoint, "/"))
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

func Do(client *http.Client, req *http.Request, v interface{}) (*Response, error) {
	rsp, err := client.Do(req)
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
