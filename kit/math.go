package kit

import "math"

// RoundToDecimal rounds a float64 to a given number of decimal places.
func RoundToDecimal(num float64, places int) float64 {
	pow := math.Pow(10, float64(places))
	return math.Round(num*pow) / pow
}
