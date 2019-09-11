package subcmd

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"time"

	"gopkg.in/src-d/go-log.v1"
)

func setLogTransport(client *http.Client, logger log.Logger) {
	t := &logTransport{client.Transport, logger}
	client.Transport = t
}

type logTransport struct {
	T      http.RoundTripper
	Logger log.Logger
}

func (t *logTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	t0 := time.Now()

	resp, err := t.T.RoundTrip(r)
	if err != nil {
		return resp, err
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return resp, err
	}

	t.Logger.With(
		log.Fields{"elapsed": time.Since(t0), "code": resp.StatusCode, "url": r.URL, "body": string(b)},
	).Debugf("HTTP response")

	resp.Body = ioutil.NopCloser(bytes.NewBuffer(b))

	return resp, err
}
