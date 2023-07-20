package main

import (
	"context"
	"flag"
	"os"
	"time"

	"github.com/jonsabados/mta2furious/mta"
	"github.com/rs/zerolog"
)

func main() {
	ctx := context.Background()

	logLevelStr := os.Getenv("LOG_LEVEL")
	if logLevelStr == "" {
		logLevelStr = "debug"
	}
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	logLevel, err := zerolog.ParseLevel(logLevelStr)
	if err != nil {
		logger.Fatal().Err(err).Msg("invalid log level")
	}
	logger = logger.Level(logLevel)

	ctx = logger.WithContext(ctx)
	apiKey := os.Getenv("MTA_API_KEY")
	var out string
	flag.StringVar(&out, "out", "output.json", "output target")
	var refreshRate time.Duration
	flag.DurationVar(&refreshRate, "refresh", time.Second*30, "refresh duration")
	flag.Parse()

	afeed := mta.NewLiveFeed("https://api-endpoint.mta.info/Dataservice/mtagtfsfeeds/nyct%2Fgtfs-ace", apiKey)
	bfeed := mta.NewLiveFeed("https://api-endpoint.mta.info/Dataservice/mtagtfsfeeds/nyct%2Fgtfs-bdfm", apiKey)
	gfeed := mta.NewLiveFeed("https://api-endpoint.mta.info/Dataservice/mtagtfsfeeds/nyct%2Fgtfs-g", apiKey)
	jfeed := mta.NewLiveFeed("https://api-endpoint.mta.info/Dataservice/mtagtfsfeeds/nyct%2Fgtfs-jz", apiKey)
	nfeed := mta.NewLiveFeed("https://api-endpoint.mta.info/Dataservice/mtagtfsfeeds/nyct%2Fgtfs-nqrw", apiKey)
	lfeed := mta.NewLiveFeed("https://api-endpoint.mta.info/Dataservice/mtagtfsfeeds/nyct%2Fgtfs-l", apiKey)
	numberedFeed := mta.NewLiveFeed("https://api-endpoint.mta.info/Dataservice/mtagtfsfeeds/nyct%2Fgtfs", apiKey)
	sfeed := mta.NewLiveFeed("https://api-endpoint.mta.info/Dataservice/mtagtfsfeeds/nyct%2Fgtfs-si", apiKey)

	transitSystem := mta.NewTransitSystem(afeed, bfeed, gfeed, jfeed, nfeed, lfeed, numberedFeed, sfeed)
	store := mta.NewMemoryStore()
	processor := mta.NewStateProcessor(transitSystem, store)

	ticker := time.Tick(refreshRate)
	_, err = processor.ProcessUpdates(ctx)
	if err != nil {
		logger.Err(err).Msg("error encountered on initial pull")
	}
	for range ticker {
		result, err := processor.ProcessUpdates(ctx)
		if err != nil {
			logger.Err(err).Msg("error encountered")
			continue
		}
		for _, segment := range result.CompletedSegments {
			logger.Info().Interface("segment", segment).Msg("a segment completed")
		}
	}
}
