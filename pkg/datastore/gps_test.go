package datastore

import "testing"

func TestNormalizeGPSPreservesValidCoordinates(t *testing.T) {
	lat, lng := NormalizeGPS(51.5074, -0.1278)
	if lat != 51.5074 || lng != -0.1278 {
		t.Fatalf("valid coordinates changed to %f, %f", lat, lng)
	}
}

func TestNormalizeGPSConstrainsOutOfRangeCoordinates(t *testing.T) {
	lat, lng := NormalizeGPS(100, 190)
	if lat < -LatMax || lat > LatMax || lng < -LngMax || lng >= LngMax {
		t.Fatalf("coordinates remain out of range: %f, %f", lat, lng)
	}
}
