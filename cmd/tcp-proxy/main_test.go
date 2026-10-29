package main

import (
	"io"
	"strings"
	"testing"
	"time"

	"github.com/juju/ratelimit"
)

type rateLimitClock struct{}

func (rateLimitClock) Now() time.Time { return time.Unix(0, 0) }
func (rateLimitClock) Sleep(time.Duration) {
	panic("unexpected wait for rate limit")
}

func TestRateLimits(t *testing.T) {
	for _, tc := range []struct {
		name                     string
		shared, upload, download string
	}{
		{name: "unlimited"},
		{name: "shared", shared: "1KB"},
		{name: "upload", upload: "2KB"},
		{name: "download", download: "3KB"},
		{name: "independent", upload: "2KB", download: "3KB"},
		{name: "combined", shared: "1KB", upload: "2KB", download: "3KB"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			limits, err := newRateLimits(tc.shared, tc.upload, tc.download)
			if err != nil {
				t.Fatal(err)
			}
			for _, bucket := range []**ratelimit.Bucket{&limits.shared, &limits.upload, &limits.download} {
				if *bucket != nil {
					capacity := (*bucket).Capacity()
					*bucket = ratelimit.NewBucketWithRateAndClock(float64(capacity), capacity, rateLimitClock{})
				}
			}
			upload := strings.NewReader("up")
			download := strings.NewReader("down")
			upReader, downReader := limits.readers(upload, download)
			for _, direction := range []struct {
				reader io.Reader
				want   string
			}{
				{upReader, "up"},
				{downReader, "down"},
			} {
				data, err := io.ReadAll(direction.reader)
				if err != nil || string(data) != direction.want {
					t.Fatalf("read = %q, %v; want %q", data, err, direction.want)
				}
			}
			if limits.shared != nil && limits.shared.Available() != 1000-6 {
				t.Fatal("shared bucket must account for both directions")
			}
			if limits.upload != nil && limits.upload.Available() != 2000-2 {
				t.Fatal("upload bucket must only account for uploaded bytes")
			}
			if limits.download != nil && limits.download.Available() != 3000-4 {
				t.Fatal("download bucket must only account for downloaded bytes")
			}
			if (limits.shared == nil) != (tc.shared == "") ||
				(limits.upload == nil) != (tc.upload == "") ||
				(limits.download == nil) != (tc.download == "") {
				t.Fatal("only configured limits should have buckets")
			}
		})
	}
}

func TestInvalidRateLimits(t *testing.T) {
	for _, flag := range []string{"rate-limit", "rate-up", "rate-down"} {
		for _, value := range []string{"invalid", "0", "-1KB"} {
			t.Run(flag+"/"+value, func(t *testing.T) {
				cmd := NewRootCmd()
				cmd.SetArgs([]string{"example.com", "8080:80", "--" + flag + "=" + value})
				if err := cmd.Execute(); err == nil || !strings.Contains(err.Error(), "--"+flag) {
					t.Fatalf("expected error naming --%s, got %v", flag, err)
				}
			})
		}
	}
}

func TestHumanSize(t *testing.T) {
	testcases := []struct {
		n    int64
		want string
	}{
		{1 << 22, "4MB"},
		{1<<11 + 110, "2.11KB"},
		{1<<11 + 119, "2.11KB"},
		{1<<11 + 99, "2KB"},
	}
	for _, tc := range testcases {
		hs := humanSize(tc.n)
		if tc.want != hs {
			t.Fatalf("unmatched values: %q != %q", tc.want, hs)
		}
	}
}
