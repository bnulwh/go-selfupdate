package selfupdate

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

const (
	DefaultToken = "ap_pJSFC5wQYkAyI0FIVwKYs9h1hW"
)

// Requester interface allows developers to customize the method in which
// requests are made to retrieve the version and binary.
type Requester interface {
	Fetch(url string) (io.ReadCloser, error)
}

// HTTPRequester is the normal requester that is used and does an HTTP
// to the URL location requested to retrieve the specified data.
type HTTPRequester struct{}

// Fetch will return an HTTP request to the specified url and return
// the body of the result. An error will occur for a non 200 status code.
func (httpRequester *HTTPRequester) Fetch(url string) (io.ReadCloser, error) {
	//fmt.Println("fetch", url)
	body := ""
	req, err := http.NewRequest("GET", url, strings.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", DefaultToken))
	client := http.DefaultClient
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("bad http status from %s: %v", url, resp.Status)
	}

	return resp.Body, nil
}

// mockRequester used for some mock testing to ensure the requester contract
// works as specified.
type mockRequester struct {
	currentIndex int
	fetches      []func(string) (io.ReadCloser, error)
}

func (mr *mockRequester) handleRequest(requestHandler func(string) (io.ReadCloser, error)) {
	if mr.fetches == nil {
		mr.fetches = []func(string) (io.ReadCloser, error){}
	}
	mr.fetches = append(mr.fetches, requestHandler)
}

func (mr *mockRequester) Fetch(url string) (io.ReadCloser, error) {
	if len(mr.fetches) <= mr.currentIndex {
		return nil, fmt.Errorf("no for currentIndex %d to mock", mr.currentIndex)
	}
	current := mr.fetches[mr.currentIndex]
	mr.currentIndex++

	return current(url)
}
