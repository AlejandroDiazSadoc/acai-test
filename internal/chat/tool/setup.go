package tool

func SetupTools() Map {
	return Map{
		"get_weather":       NewWeatherTool(),
		"get_today_date":    NewDateTool(),
		"get_holidays":      NewHolidaysTool(),
		"get_flight_status": NewFlightStatusTool(),
	}
}
