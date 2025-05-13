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
	client          *http.Client
	initOnce        sync.Once
	proxyAddrGlobal string
	insecureGlobal  bool
)

func Init(proxyAddr string, insecure bool) error {
	var err error
	proxyAddrGlobal = proxyAddr
	insecureGlobal = insecure

	initOnce.Do(func() {
		client = &http.Client{
			Timeout: 30 * time.Second,
		}
		err = ResetTransport(proxyAddr, insecure)
	})

	return err
}

func Get() *http.Client {
	return client
}

func ResetTransport(proxyAddr string, insecure bool) error {
	transport := &http.Transport{
		TLSClientConfig:     &tls.Config{InsecureSkipVerify: insecure},
		DisableKeepAlives:   true,
		MaxIdleConns:        0,
		MaxIdleConnsPerHost: 0,
		IdleConnTimeout:     0,
		TLSNextProto:        make(map[string]func(string, *tls.Conn) http.RoundTripper),
	}

	if proxyAddr != "" {
		if strings.HasPrefix(proxyAddr, "socks5://") {
			parsed, _ := url.Parse(proxyAddr)
			dialer, err := proxy.SOCKS5("tcp", parsed.Host, nil, proxy.Direct)
			if err != nil {
				return err
			}
			transport.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
				return dialer.Dial(network, addr)
			}
		} else {
			transport.Proxy = func(*http.Request) (*url.URL, error) {
				return url.Parse(proxyAddr)
			}
		}
	}

	if client != nil {
		if oldTransport, ok := client.Transport.(*http.Transport); ok {
			oldTransport.CloseIdleConnections()
		}
		client.Transport = transport
	}

	return nil
}

func Reset() error {
	return ResetTransport(proxyAddrGlobal, insecureGlobal)
}
