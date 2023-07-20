package static

import (
	"os"

	csv "github.com/trimmer-io/go-csv"
)

type Route struct {
	RouteID     string `csv:"route_id"`
	ShortName   string `csv:"route_short_name"`
	LongName    string `csv:"route_long_name"`
	Description string `csv:"route_desc"`
	Type        int    `csv:"route_type"`
	Color       string `csv:"route_color"`
}

type Station struct {
	StationID int     `csv:"Station ID"`
	GTFSID    string  `csv:"GTFS Stop ID"`
	Line      string  `csv:"Line"`
	Name      string  `csv:"Stop Name"`
	Borough   string  `csv:"Borough"`
	Structure string  `csv:"Structure"`
	Latitude  float64 `csv:"GTFS Latitude"`
	Longitude float64 `csv:"GTFS Longitude"`
}

type StopTime struct {
	TripID       string `csv:"trip_id"`
	StopID       string `csv:"stop_id"`
	StopSequence int    `csv:"stop_sequence"`
}

type Transfer struct {
	FromStopID          string `csv:"from_stop_id"`
	ToStopID            string `csv:"to_stop_id"`
	TransferType        int    `csv:"transfer_type"`
	MinTransferTimeSecs int    `csv:"min_transfer_time"`
}

// MustLoad populates the slice of type T with the data contained in the file at prefix + filename,
// returning error if any errors are encountered. See go-csv docs for 'csv' tag details
func MustLoad[T any](prefix, filename string, dest *[]T) {
	b, err := os.ReadFile(prefix + filename)
	if err != nil {
		panic(err)
	}

	if err := csv.Unmarshal(b, dest); err != nil {
		panic(err)
	}
}
