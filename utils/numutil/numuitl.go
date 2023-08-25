package numutil

import "math"

func TruncateFloat64(float float64, decimal int) float64 {
	shift := math.Pow(10, float64(decimal))
	return math.Round(float*shift) / shift
}
