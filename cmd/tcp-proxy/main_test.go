package main

import "testing"

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
