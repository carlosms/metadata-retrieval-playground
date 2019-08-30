package client

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gregjones/httpcache"
	"github.com/gregjones/httpcache/diskcache"
	"github.com/src-d/ghsync/utils"
)

func NewClient(httpClient *http.Client) (*http.Client, error) {
	dirPath := filepath.Join(os.TempDir(), "ghsync")
	err := os.MkdirAll(dirPath, os.ModePerm)
	if err != nil {
		return nil, fmt.Errorf("error while creating directory %s: %v", dirPath, err)
	}

	t := httpcache.NewTransport(diskcache.New(dirPath))
	t.Transport = &RemoveHeaderTransport{
		T: utils.NewRateLimitTransport(httpClient.Transport),
	}
	httpClient.Transport = &RetryTransport{T: t}

	return httpClient, nil
}

type RemoveHeaderTransport struct {
	T http.RoundTripper
}

func (t *RemoveHeaderTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Del("X-Ratelimit-Limit")
	req.Header.Del("X-Ratelimit-Remaining")
	req.Header.Del("X-Ratelimit-Reset")
	return t.T.RoundTrip(req)
}

type RetryTransport struct {
	T http.RoundTripper
}

func (t *RetryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	var r *http.Response
	var err error
	utils.Retry(func() error {
		r, err = t.T.RoundTrip(req)
		return err
	})

	return r, err
}
