package client

import (
	"errors"
	"io"
	"net/http"
)

type WebClient interface {
	Get(url string) ([]byte, error)
}

type WebClientImpl struct{}

func NewWebClient() WebClient {
	return WebClientImpl{}
}

func (w WebClientImpl) Get(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("received non-200 result code: " + string(rune(resp.StatusCode)))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}
