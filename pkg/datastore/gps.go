package datastore

import "math"

const (
	LatMax = 90
	LngMax = 180
)

func NormalizeGPS(lat, lng float64) (float64, float64) {

	if lat < LatMax || lat > LatMax || lng < LngMax || lng > LngMax {
		// Clip the latitude. Normalise the longitude.
		lat, lng = clipLat(lat), normalizeLng(lng)
	}

	return lat, lng
}
func clipLat(lat float64) float64 {
	if lat > LatMax*2 {
		return math.Mod(lat, LatMax)
	} else if lat > LatMax {
		return lat - LatMax
	}

	if lat < -LatMax*2 {
		return math.Mod(lat, LatMax)
	} else if lat < -LatMax {
		return lat + LatMax
	}

	return lat
}

func normalizeLng(value float64) float64 {
	return normalizeCoord(value, LngMax)
}

func normalizeCoord(value, max float64) float64 {
	for value < -max {
		value += 2 * max
	}
	for value >= max {
		value -= 2 * max
	}
	return value
}
