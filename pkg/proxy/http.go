package proxy

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/eirture/tcp-proxy/pkg/build"
	"golang.org/x/net/context"
)

var (
	KeepAliveTime = 180 * time.Second
	// DialTimeout is the timeout of dial.
	DialTimeout    = 5 * time.Second
	ConnectTimeout = 5 * time.Second
)

type HTTPOptions struct {
	UserAgent string
	User      *url.Userinfo
	Timeout   time.Duration
}

type HTTPDialer struct {
	forward DialerContext
	address string
	options HTTPOptions
}

func NewHTTPDialer(addr string, forward Dialer, options HTTPOptions) DialerContext {
	return &HTTPDialer{
		forward: DialerContextFunc(func(ctx context.Context, network, addr string) (net.Conn, error) {
			if d, ok := forward.(DialerContext); ok {
				return d.DialContext(ctx, network, addr)
			}
			return forward.Dial(network, addr)
		}),
		address: addr,
		options: options,
	}
}

func (d *HTTPDialer) Dial(network, addr string) (net.Conn, error) {
	return d.DialContext(context.Background(), network, addr)
}

func (d *HTTPDialer) DialContext(ctx context.Context, network, addr string) (conn net.Conn, err error) {
	if network != "tcp" {
		return nil, fmt.Errorf("http proxy: %s unsupported", network)
	}

	conn, err = d.forward.DialContext(ctx, network, d.address)
	if err != nil {
		return
	}
	defer func() {
		if err != nil {
			conn.Close()
		}
	}()

	timeout := d.options.Timeout
	if timeout <= 0 {
		timeout = ConnectTimeout
	}
	conn.SetDeadline(time.Now().Add(timeout))
	defer conn.SetDeadline(time.Time{})
	req := &http.Request{
		Method:     http.MethodConnect,
		URL:        &url.URL{Host: addr},
		Host:       addr,
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     make(http.Header),
	}
	userAgent := "tcp-proxy/" + build.VersionWithDate()
	if d.options.UserAgent != "" {
		userAgent = d.options.UserAgent
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Proxy-Connection", "keep-alive")

	user := d.options.User
	if user != nil {
		u := user.Username()
		p, _ := user.Password()
		req.Header.Set("Proxy-Authorization",
			"Basic "+base64.StdEncoding.EncodeToString([]byte(u+":"+p)))
	}

	if err = req.Write(conn); err != nil {
		return nil, err
	}

	resp, err := http.ReadResponse(bufio.NewReader(conn), req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s", resp.Status)
	}

	return conn, nil
}
