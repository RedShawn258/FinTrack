package models

import "time"

// ForecastRequest represents a request for expense forecasting
type ForecastRequest struct {
	CategoryID  *uint      `json:"categoryId,omitempty"` // Optional: filter by category
	MonthsAhead int        `json:"monthsAhead"`          // How many months to forecast
	StartDate   *time.Time `json:"startDate,omitempty"`  // Optional: custom start date
}

// ForecastPoint represents a single point in the forecast
type ForecastPoint struct {
	Month       string  `json:"month"`       // Month in YYYY-MM format
	Amount      float64 `json:"amount"`      // Forecasted amount
	Probability float64 `json:"probability"` // Confidence level (0-1)
}

// CategoryForecast represents the forecast for a specific category
type CategoryForecast struct {
	CategoryID   uint            `json:"categoryId"`
	CategoryName string          `json:"categoryName"`
	Forecast     []ForecastPoint `json:"forecast"`
}

// ForecastResponse represents the full response to a forecast request
type ForecastResponse struct {
	TotalForecast     []ForecastPoint    `json:"totalForecast"`     // Overall forecast
	CategoryForecasts []CategoryForecast `json:"categoryForecasts"` // Category-specific forecasts
	Message           string             `json:"message"`           // Additional info about the forecast
}
