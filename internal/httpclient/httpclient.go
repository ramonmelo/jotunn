package httpclient

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/proxy"
)

var (
	client   *http.Client
	initOnce sync.Once
)

func Init(proxyAddr string, insecure bool) error {
	var err error

	initOnce.Do(func() {
		transport := &http.Transport{
			TLSClientConfig:     &tls.Config{InsecureSkipVerify: insecure},
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 100,
			IdleConnTimeout:     90 * time.Second,
		}

		if proxyAddr != "" {
			if strings.HasPrefix(proxyAddr, "socks5://") {
				parsed, _ := url.Parse(proxyAddr)
				dialer, derr := proxy.SOCKS5("tcp", parsed.Host, nil, proxy.Direct)
				if derr != nil {
					err = derr
					return
				}

				transport.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
					return dialer.Dial(network, addr)
				}
			} else {
				proxyFunc := func(*http.Request) (*url.URL, error) {
					return url.Parse(proxyAddr)
				}
				transport.Proxy = proxyFunc
			}
		}

		client = &http.Client{
			Transport: transport,
			Timeout:   30 * time.Second,
		}
	})

	return err
}

func Get() *http.Client {
	return client
}
