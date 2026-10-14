package datastore

import "testing"

func TestShutterSpeedDisplay(t *testing.T) {
	for _, tc := range []struct{ raw, want string }{
		{"24999979/1000000000", "1/40 s"},
		{"1/125", "1/125 s"},
		{"0.008", "1/125 s"},
		{"1/2", "0.5 s"},
		{"3/2", "1.5 s"},
		{"30/1", "30 s"},
		{"", ""}, {"1/0", "1/0"}, {"0", "0"}, {"Unknown", "Unknown"},
	} {
		t.Run(tc.raw, func(t *testing.T) {
			if got := (Picture{ShutterSpeed: tc.raw}).ShutterSpeedDisplay(); got != tc.want {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
		})
	}
}
