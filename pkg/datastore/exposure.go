package datastore

import (
	"fmt"
	"math"
	"math/big"
	"strconv"
	"strings"
)

// ShutterSpeedDisplay formats existing EXIF values without requiring a rescan.
func (p Picture) ShutterSpeedDisplay() string {
	raw := strings.TrimSpace(p.ShutterSpeed)
	value, ok := new(big.Rat).SetString(raw)
	if !ok {
		return raw
	}
	seconds, _ := value.Float64()
	if seconds <= 0 || math.IsInf(seconds, 0) {
		return raw
	}
	if seconds < 0.5 {
		return fmt.Sprintf("1/%.0f s", math.Round(1/seconds))
	}
	return strconv.FormatFloat(seconds, 'g', 3, 64) + " s"
}
