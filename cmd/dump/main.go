package main

import (
	"context"
	"encoding/json"
	"flag"
	"os"

	"github.com/jonsabados/mta2furious/mta"
)

func main() {
	ctx := context.Background()
	apiKey := os.Getenv("MTA_API_KEY")
	var out string
	flag.StringVar(&out, "out", "output.json", "output target")
	flag.Parse()
	feed := mta.NewFeed("https://api-endpoint.mta.info/Dataservice/mtagtfsfeeds/nyct%2Fgtfs-ace", apiKey)
	currentState, err := feed.CurrentFeed(ctx)
	if err != nil {
		panic(err)
	}
	jsonBytes, err := json.Marshal(currentState)
	if err != nil {
		panic(err)
	}
	err = os.WriteFile(out, jsonBytes, 0644)
	if err != nil {
		panic(err)
	}
}
