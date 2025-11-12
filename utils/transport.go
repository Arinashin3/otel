package utils

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
)

func NewTransport(insecure bool) *http.Transport {
	return &http.Transport{
		MaxConnsPerHost: 1,
		DialTLSContext: func(ctx context.Context, network string, addr string) (net.Conn, error) {
			return tls.Dial(network, addr, &tls.Config{InsecureSkipVerify: insecure})
		},
	}

}
