package common

import (
	"crypto/tls"
	"fmt"
	"github.com/orange-cloudfoundry/statusetat/models"
	"io/ioutil"
	"net"
	"net/http"
	"time"
)

type HeaderTransport struct {
	key           string
	val           string
	WrapTransport http.RoundTripper
}

func NewHeaderTransport(transport http.RoundTripper, key, val string) *HeaderTransport {
	return &HeaderTransport{
		key:           key,
		val:           val,
		WrapTransport: transport,
	}
}

func (t HeaderTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set(t.key, t.val)
	return t.WrapTransport.RoundTrip(req)
}

func MakeHttpTransportWithHeader(skipInsecure bool, key, val string) http.RoundTripper {
	return NewHeaderTransport(MakeHttpTransport(skipInsecure), key, val)
}

func ExtractHttpError(resp *http.Response) error {
	if resp.StatusCode > 399 {
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("Get error code %d", resp.StatusCode)
		}
		return fmt.Errorf("Get error code %d: %s", resp.StatusCode, string(b))
	}
	return nil
}

func MakeHttpTransport(skipInsecure bool) http.RoundTripper {
	return &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: skipInsecure,
		},
	}
}

func MetadataToMap(metadata []models.Metadata) map[string]string {
	m := make(map[string]string)
	for _, elem := range metadata {
		m[elem.Key] = elem.Value
	}
	return m
}
