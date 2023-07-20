package static

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const fixtures = "./fixtures/"

func TestLoad_Routes(t *testing.T) {
	exp := []Route{{
		RouteID:     "1",
		ShortName:   "1",
		LongName:    "Broadway - 7 Avenue Local",
		Description: "Trains operate between 242 St in the Bronx and South Ferry in Manhattan, at all times",
		Type:        1,
		Color:       "EE352E",
	}, {
		RouteID:     "5",
		ShortName:   "5",
		LongName:    "Lexington Avenue Express",
		Description: "Weekdays daytime, most trains operate between either Dyre Av or 238 St-Nereid Av, Bronx, and Flatbush Av-Brooklyn College, Brooklyn. At all other times except during late nights, trains operate between Dyre Av, Bronx, and Bowling Green, Manhattan. During late nights trains operate only in the Bronx between Dyre Av and E 180 St/MorrisPark Av. Customers who ride during late night hours can transfer to 2 service at the E 180 St Station. At all times, trains operate express in Manhattan and Brooklyn. Weekdays, trains in the Bronx operate express from E 180 St to 149 St-3 Av during morning rush hours (from about 6 AM to 9 AM), and from 149 St-3 Av to E 180 St during the evening rush hours (from about 4 PM to 7 PM).",
		Type:        1,
		Color:       "00933C",
	}, {
		RouteID:     "5X",
		ShortName:   "5X",
		LongName:    "Lexington Avenue Express",
		Description: "Weekdays daytime, most trains operate between either Dyre Av or 238 St-Nereid Av, Bronx, and Flatbush Av-Brooklyn College, Brooklyn. At all other times except during late nights, trains operate between Dyre Av, Bronx, and Bowling Green, Manhattan. During late nights trains operate only in the Bronx between Dyre Av and E 180 St/MorrisPark Av. Customers who ride during late night hours can transfer to 2 service at the E 180 St Station. At all times, trains operate express in Manhattan and Brooklyn. Weekdays, trains in the Bronx operate express from E 180 St to 149 St-3 Av during morning rush hours (from about 6 AM to 9 AM), and from 149 St-3 Av to E 180 St during the evening rush hours (from about 4 PM to 7 PM).",
		Type:        1,
		Color:       "00933C",
	}, {
		RouteID:     "A",
		ShortName:   "A",
		LongName:    "8 Avenue Express",
		Description: "Trains operate between Inwood-207 St, Manhattan and Far Rockaway-Mott Avenue, Queens at all times. Also from about 6 AM until about midnight, additional trains operate between Inwood-207 St and Lefferts Boulevard (trains typically alternate between Lefferts Blvd and Far Rockaway). During weekday morning rush hours, special trains operate from Rockaway Park-Beach 116 St, Queens, toward Manhattan. These trains make local stops between Rockaway Park and Broad Channel. Similarly, in the evening rush hour special trains leave Manhattan operating toward Rockaway Park-Beach 116 St, Queens.",
		Type:        1,
		Color:       "2850AD",
	}}

	out := make([]Route, 0)
	MustLoad[Route](fixtures, "routes.txt", &out)
	assert.EqualValues(t, exp, out)
}

func TestLoad_Stations(t *testing.T) {
	exp := []Station{{
		StationID: 1,
		GTFSID:    "R01",
		Line:      "Astoria",
		Name:      "Astoria-Ditmars Blvd",
		Borough:   "Q",
		Structure: "Elevated",
		Latitude:  40.775036,
		Longitude: -73.912034,
	}, {
		StationID: 7,
		GTFSID:    "R11",
		Line:      "Astoria",
		Name:      "Lexington Av/59 St",
		Borough:   "M",
		Structure: "Subway",
		Latitude:  40.76266,
		Longitude: -73.967258,
	}, {
		StationID: 79,
		GTFSID:    "N10",
		Line:      "Sea Beach",
		Name:      "86 St",
		Borough:   "Bk",
		Structure: "Open Cut",
		Latitude:  40.592721,
		Longitude: -73.97823,
	}, {
		StationID: 120,
		GTFSID:    "L08",
		Line:      "Canarsie",
		Name:      "Bedford Av",
		Borough:   "Bk",
		Structure: "Subway",
		Latitude:  40.717304,
		Longitude: -73.956872,
	}, {
		StationID: 143,
		GTFSID:    "A02",
		Line:      "8th Av - Fulton St",
		Name:      "Inwood-207 St",
		Borough:   "M",
		Structure: "Subway",
		Latitude:  40.868072,
		Longitude: -73.919899,
	}}

	out := make([]Station, 0)
	MustLoad[Station](fixtures, "Stations.csv", &out)
	assert.EqualValues(t, exp, out)
}

func TestLoad_StopTimes(t *testing.T) {
	exp := []StopTime{{
		TripID:       "ASP23GEN-1037-Sunday-00_000600_1..S03R",
		StopID:       "101S",
		StopSequence: 1,
	}, {
		TripID:       "ASP23GEN-1037-Sunday-00_000600_1..S03R",
		StopID:       "103S",
		StopSequence: 2,
	}, {
		TripID:       "ASP23GEN-1037-Sunday-00_000600_1..S03R",
		StopID:       "104S",
		StopSequence: 3,
	}}

	out := make([]StopTime, 0)
	MustLoad[StopTime](fixtures, "stop_times.txt", &out)
	assert.EqualValues(t, exp, out)
}

func TestLoad_Transfers(t *testing.T) {
	exp := []Transfer{{
		FromStopID:          "101",
		ToStopID:            "101",
		TransferType:        2,
		MinTransferTimeSecs: 180,
	}, {
		FromStopID:          "112",
		ToStopID:            "A09",
		TransferType:        2,
		MinTransferTimeSecs: 180,
	}, {
		FromStopID:          "123",
		ToStopID:            "123",
		TransferType:        2,
		MinTransferTimeSecs: 0,
	}}

	out := make([]Transfer, 0)
	MustLoad[Transfer](fixtures, "transfers.txt", &out)
	assert.EqualValues(t, exp, out)
}
