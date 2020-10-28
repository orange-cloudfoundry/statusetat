package storages

import (
	"crypto/tls"
	"net"
	"net/http"
	"net/url"
	"time"
)

const subscriberFilename = "subscribers.json"

func makeHttpClient(u *url.URL) *http.Client {
	transport := makeHttpTransport(u)
	client := &http.Client{
		Transport: transport,
		Timeout:   0,
	}
	return client
}

func makeHttpTransport(u *url.URL) *http.Transport {
	skipInsecure := false
	if u.Query().Get("insecure-skip-verify") != "" {
		skipInsecure = true
		val := u.Query()
		val.Del("insecure-skip-verify")
		u.RawQuery = val.Encode()
	}
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
