package models

type FlightStatusResponse struct {
	Data []DatedFlight `json:"data"`
}

// DatedFlight contains all details for a specific flight on a specific day.
type DatedFlight struct {
	Type                   string           `json:"type"`                   // e.g., "DatedFlight"
	ScheduledDepartureDate string           `json:"scheduledDepartureDate"` // e.g., "2023-08-01"
	FlightDesignator       FlightDesignator `json:"flightDesignator"`
	FlightPoints           []FlightPoint    `json:"flightPoints"`
	Segments               []Segment        `json:"segments"`
	// Legs is also present but often contains redundant or highly specific data.
	// Including it for completeness:
	Legs []Leg `json:"legs"`
}

// FlightDesignator identifies the marketed flight.
type FlightDesignator struct {
	CarrierCode  string `json:"carrierCode"`  // e.g., "TP"
	FlightNumber int    `json:"flightNumber"` // e.g., 487
}

// FlightPoint represents a point in the flight journey (departure or arrival airport).
type FlightPoint struct {
	IataCode  string          `json:"iataCode"` // e.g., "NCE", "LIS"
	Departure *DeparturePoint `json:"departure,omitempty"`
	Arrival   *ArrivalPoint   `json:"arrival,omitempty"`
}

// DeparturePoint contains departure-specific information.
type DeparturePoint struct {
	Timings []Timing `json:"timings"`
}

// ArrivalPoint contains arrival-specific information.
type ArrivalPoint struct {
	Timings []Timing `json:"timings"`
}

// Timing contains the time information and its qualifier (STD/STA/ATD/ATA).
type Timing struct {
	Qualifier string `json:"qualifier"` // e.g., "STD" (Scheduled Time of Departure)
	Value     string `json:"value"`     // e.g., "2023-08-01T18:10+02:00"
}

// Segment represents a flight segment with operating carrier details.
type Segment struct {
	BoardPointIataCode       string      `json:"boardPointIataCode"`       // e.g., "NCE"
	OffPointIataCode         string      `json:"offPointIataCode"`         // e.g., "LIS"
	ScheduledSegmentDuration string      `json:"scheduledSegmentDuration"` // e.g., "PT2H35M" (ISO 8601 Duration)
	Partnership              Partnership `json:"partnership"`
}

// Partnership holds information about the operating carrier (for code-shares).
type Partnership struct {
	OperatingFlight OperatingFlight `json:"operatingFlight"`
}

// OperatingFlight identifies the aircraft's actual operator.
type OperatingFlight struct {
	CarrierCode  string `json:"carrierCode"`  // e.g., "A3"
	FlightNumber int    `json:"flightNumber"` // e.g., 1748
}

// Leg provides technical details about the physical flight leg.
type Leg struct {
	BoardPointIataCode   string            `json:"boardPointIataCode"`
	OffPointIataCode     string            `json:"offPointIataCode"`
	AircraftEquipment    AircraftEquipment `json:"aircraftEquipment"`
	ScheduledLegDuration string            `json:"scheduledLegDuration"`
}

// AircraftEquipment specifies the type of aircraft.
type AircraftEquipment struct {
	AircraftType string `json:"aircraftType"` // e.g., "E90"
}
