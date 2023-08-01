package proxy

import (
	"net"
	"net/url"

	"golang.org/x/net/context"
	"golang.org/x/net/proxy"
)

func init() {
	proxy.RegisterDialerType("http", func(u *url.URL, forward proxy.Dialer) (proxy.Dialer, error) {
		return NewHTTPDialer(u.Host, forward, HTTPOptions{User: u.User}), nil
	})
}

var (
	Direct Dialer = proxy.Direct
)

type Dialer interface {
	Dial(network, addr string) (c net.Conn, err error)
}

type DialerContext interface {
	Dialer
	DialContext(ctx context.Context, network, addr string) (net.Conn, error)
}

type DialerContextFunc func(ctx context.Context, network, addr string) (net.Conn, error)

func (f DialerContextFunc) Dial(network, addr string) (net.Conn, error) {
	return f(context.Background(), network, addr)
}

func (f DialerContextFunc) DialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	return f(ctx, network, addr)
}

func NewDialer(rawURL string, forward Dialer) (Dialer, error) {
	if rawURL == "" {
		return proxy.FromEnvironment(), nil
	}
	proxyUrl, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}
	return proxy.FromURL(proxyUrl, forward)
}
