package models

type WeatherAPIResponse struct {
	Location Location `json:"location"`
	Current  Current  `json:"current"`
	Forecast Forecast `json:"forecast"`
}

type Location struct {
	Name    string `json:"name"`
	Region  string `json:"region"`
	Country string `json:"country"`
}

type Condition struct {
	Text string `json:"text"`
}

type Current struct {
	TempC     float64   `json:"temp_c"`
	WindKph   float64   `json:"wind_kph"`
	Condition Condition `json:"condition"`
}

type Forecast struct {
	ForecastDay []ForecastDay `json:"forecastday"`
}

type ForecastDay struct {
	Date string `json:"date"`
	Day  Day    `json:"day"`
}

type Day struct {
	MaxtempC  float64   `json:"maxtemp_c"`
	MintempC  float64   `json:"mintemp_c"`
	Condition Condition `json:"condition"`
}
