package proxy

import "testing"

func TestNewDialer(t *testing.T) {
	testcases := []struct {
		rawURL string
		isHTTP bool
	}{
		{"127.0.0.1:8080", true},
		{"user:pass@127.0.0.1:8080", true},
		{"http://127.0.0.1:8080", true},
		{"socks5://127.0.0.1:1080", false},
	}
	for _, tc := range testcases {
		d, err := NewDialer(tc.rawURL, Direct)
		if err != nil {
			t.Fatalf("NewDialer(%q) error: %v", tc.rawURL, err)
		}
		if _, ok := d.(*HTTPDialer); ok != tc.isHTTP {
			t.Fatalf("NewDialer(%q) returned %T", tc.rawURL, d)
		}
	}
}
