package utils

// CentsToDollars converts cents (int64) to dollars (float64) for API responses.
// This is a temporary measure until we fully migrate to integer-based money handling.
func CentsToDollars(cents int64) float64 {
	return float64(cents) / 100.0
}

// DollarsToCents converts dollars (float64) to cents (int64) for internal calculations.
// Rounds to nearest cent to handle floating point imprecision.
func DollarsToCents(dollars float64) int64 {
	// Multiply by 100 and round to nearest integer
	return int64(dollars*100 + 0.5)
}
