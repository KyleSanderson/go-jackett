package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/autobrr/go-jackett"
)

func main() {
	// Example 1: Using Jackett proxy mode (original behavior)
	jackettClient := jackett.NewClient(jackett.Config{
		Host:   "http://localhost:9117",
		APIKey: "your-jackett-api-key",
	})

	// Search via Jackett proxy
	opts := map[string]string{
		"t": "search",
		"q": "ubuntu",
	}
	results, err := jackettClient.GetTorrents("all", opts)
	if err != nil {
		log.Printf("Jackett proxy search error: %v", err)
	} else {
		fmt.Printf("Found %d results via Jackett\n", len(results.Channel.Items))
	}

	// Example 2: Using direct tracker mode
	directClient := jackett.NewClient(jackett.Config{
		Host:       "https://www.morethantv.me/api/torznab",
		APIKey:     "your-tracker-api-key",
		DirectMode: true,
		Timeout:    30,
	})

	// Search directly on the tracker
	directResults, err := directClient.SearchDirect("ubuntu", map[string]string{
		"limit": "50",
		"cat":   "5000", // Example category
	})
	if err != nil {
		log.Printf("Direct search error: %v", err)
	} else {
		fmt.Printf("Found %d results directly from tracker\n", len(directResults.Channel.Items))
	}

	// Get tracker capabilities
	caps, err := directClient.GetCapsDirect()
	if err != nil {
		log.Printf("Failed to get capabilities: %v", err)
	} else {
		fmt.Printf("Tracker capabilities retrieved\n")
		for _, indexer := range caps.Indexer {
			fmt.Printf("Indexer: %s\n", indexer.Title)
		}
	}

	// Example 3: Using context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	timedResults, err := directClient.SearchDirectCtx(ctx, "debian", nil)
	if err != nil {
		log.Printf("Timed search error: %v", err)
	} else {
		fmt.Printf("Found %d results with timeout context\n", len(timedResults.Channel.Items))
	}
}
