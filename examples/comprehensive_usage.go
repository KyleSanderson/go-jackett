package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/autobrr/go-jackett"
)

func main() {
	// Example configurations
	exampleJackettProxy()
	exampleDirectTracker()
	exampleTVSearch()
	exampleMovieSearch()
	exampleMusicSearch()
	exampleBookSearch()
	exampleTorznabHelpers()
}

// Example 1: Using Jackett as a proxy to multiple indexers
func exampleJackettProxy() {
	fmt.Println("=== Jackett Proxy Example ===")

	client := jackett.NewClient(jackett.Config{
		Host:   "http://localhost:9117",
		APIKey: "your-jackett-api-key",
	})

	// Basic search
	opts := map[string]string{
		"t":   "search",
		"q":   "ubuntu",
		"cat": jackett.CategoryAll,
	}

	results, err := client.GetTorrents("all", opts)
	if err != nil {
		log.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Found %d results\n\n", len(results.Channel.Item))
}

// Example 2: Direct tracker access (bypass Jackett)
func exampleDirectTracker() {
	fmt.Println("=== Direct Tracker Example ===")

	client := jackett.NewClient(jackett.Config{
		Host:       "https://www.morethantv.me/api/torznab",
		APIKey:     "your-tracker-api-key",
		DirectMode: true,
		Timeout:    30,
	})

	// Get tracker capabilities first
	caps, err := client.GetCapsDirect()
	if err != nil {
		log.Printf("Error getting caps: %v\n", err)
		return
	}

	fmt.Printf("Tracker capabilities retrieved\n")
	for _, indexer := range caps.Indexer {
		fmt.Printf("  Indexer: %s\n", indexer.Title)
	}

	// Direct search
	results, err := client.SearchDirect("ubuntu", map[string]string{
		"limit": "10",
		"cat":   jackett.CategoryTVHD,
	})
	if err != nil {
		log.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Found %d results\n\n", len(results.Channel.Item))
}

// Example 3: TV show search with specific parameters
func exampleTVSearch() {
	fmt.Println("=== TV Search Example ===")

	client := jackett.NewClient(jackett.Config{
		Host:   "http://localhost:9117",
		APIKey: "your-api-key",
	})

	// Search for a specific TV show by TVDB ID
	results, err := client.TVSearch(jackett.TVSearchOptions{
		Query:    "The Expanse",
		TVDBID:   "280619",
		Season:   "6",
		Episode:  "5",
		Category: jackett.CategoryTVHD,
		Limit:    "50",
	})
	if err != nil {
		log.Printf("Error: %v\n", err)
		return
	}

	items := results.ToTorznabItems()
	fmt.Printf("Found %d results for The Expanse S06E05\n", len(items))

	for _, item := range items {
		fmt.Printf("  Title: %s\n", item.Title)
		fmt.Printf("  Seeders: %d, Leechers: %d\n", item.Seeders(), item.Leechers())
		fmt.Printf("  TVDB: %s, Season: %d, Episode: %d\n",
			item.TVDBID(), item.Season(), item.Episode())
		if item.IsFreeleech() {
			fmt.Printf("  ** FREELEECH **\n")
		}
		fmt.Println()
	}
	fmt.Println()
}

// Example 4: Movie search with IMDB ID
func exampleMovieSearch() {
	fmt.Println("=== Movie Search Example ===")

	client := jackett.NewClient(jackett.Config{
		Host:       "https://tracker.example.com/api/torznab",
		APIKey:     "your-tracker-api-key",
		DirectMode: true,
	})

	// Search for a movie by IMDB ID
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	results, err := client.MovieSearchCtx(ctx, jackett.MovieSearchOptions{
		IMDBID:   "tt0816692", // Interstellar
		Year:     "2014",
		Category: jackett.CategoryMoviesUHD,
		Limit:    "25",
		Extended: "1", // Get all extended attributes
	})
	if err != nil {
		log.Printf("Error: %v\n", err)
		return
	}

	items := results.ToTorznabItems()
	fmt.Printf("Found %d results for Interstellar (2014)\n", len(items))

	for _, item := range items {
		fmt.Printf("  %s\n", item.Title)
		fmt.Printf("  IMDB: %s, Year: %d\n", item.IMDBID(), item.Year())
		fmt.Printf("  Resolution: %s, Video: %s, Audio: %s\n",
			item.Resolution(), item.Video(), item.Audio())
		fmt.Printf("  Seeders: %d, Size: %s\n", item.Seeders(), item.Size)
		fmt.Println()
	}
	fmt.Println()
}

// Example 5: Music search with artist and album
func exampleMusicSearch() {
	fmt.Println("=== Music Search Example ===")

	client := jackett.NewClient(jackett.Config{
		Host:   "http://localhost:9117",
		APIKey: "your-api-key",
	})

	results, err := client.MusicSearch(jackett.MusicSearchOptions{
		Artist:   "Pink Floyd",
		Album:    "Dark Side of the Moon",
		Category: jackett.CategoryAudioLossless,
		Limit:    "20",
	})
	if err != nil {
		log.Printf("Error: %v\n", err)
		return
	}

	items := results.ToTorznabItems()
	fmt.Printf("Found %d results for Pink Floyd albums\n", len(items))

	for _, item := range items {
		fmt.Printf("  %s\n", item.Title)
		fmt.Printf("  Artist: %s, Album: %s\n", item.Artist(), item.Album())
		fmt.Printf("  Genre: %s, Year: %d\n", item.Genre(), item.Year())
		fmt.Println()
	}
	fmt.Println()
}

// Example 6: Book search with author
func exampleBookSearch() {
	fmt.Println("=== Book Search Example ===")

	client := jackett.NewClient(jackett.Config{
		Host:   "http://localhost:9117",
		APIKey: "your-api-key",
	})

	results, err := client.BookSearch(jackett.BookSearchOptions{
		Author:   "Isaac Asimov",
		Title:    "Foundation",
		Category: jackett.CategoryBooksEBook,
	})
	if err != nil {
		log.Printf("Error: %v\n", err)
		return
	}

	items := results.ToTorznabItems()
	fmt.Printf("Found %d results for Isaac Asimov books\n", len(items))

	for _, item := range items {
		fmt.Printf("  %s\n", item.Title)
		fmt.Printf("  Author: %s, Book: %s\n", item.Author(), item.BookTitle())
		fmt.Println()
	}
	fmt.Println()
}

// Example 7: Using torznab helper methods
func exampleTorznabHelpers() {
	fmt.Println("=== Torznab Helpers Example ===")

	client := jackett.NewClient(jackett.Config{
		Host:   "http://localhost:9117",
		APIKey: "your-api-key",
	})

	results, err := client.SearchDirect("ubuntu", nil)
	if err != nil {
		log.Printf("Error: %v\n", err)
		return
	}

	items := results.ToTorznabItems()

	// Filter for only freeleech torrents with good seeds
	var freeleechItems []jackett.TorznabItem
	for _, item := range items {
		if item.IsFreeleech() && item.Seeders() >= 5 {
			freeleechItems = append(freeleechItems, item)
		}
	}

	fmt.Printf("Found %d freeleech items with 5+ seeders\n", len(freeleechItems))

	for _, item := range freeleechItems {
		fmt.Printf("\n  Title: %s\n", item.Title)
		fmt.Printf("  InfoHash: %s\n", item.InfoHash())
		fmt.Printf("  Magnet: %s\n", item.MagnetURL())
		fmt.Printf("  Health: %d seeders / %d leechers\n",
			item.Seeders(), item.Leechers())
		fmt.Printf("  Download Factor: %.1f (%.0f%% off)\n",
			item.DownloadVolumeFactor(),
			(1.0-item.DownloadVolumeFactor())*100)
		fmt.Printf("  Upload Factor: %.1f\n", item.UploadVolumeFactor())

		// Seeding requirements
		if item.MinimumRatio() > 0 {
			fmt.Printf("  Min Ratio: %.2f\n", item.MinimumRatio())
		}
		if item.MinimumSeedTime() > 0 {
			hours := float64(item.MinimumSeedTime()) / 3600.0
			fmt.Printf("  Min Seed Time: %.1f hours\n", hours)
		}

		// Categories
		fmt.Printf("  Categories: %v\n", item.Categories())

		// Tags
		if len(item.Tags()) > 0 {
			fmt.Printf("  Tags: %v\n", item.Tags())
			if item.HasTag("internal") {
				fmt.Printf("  ⭐ Internal release\n")
			}
			if item.HasTag("trusted") {
				fmt.Printf("  ✓ Trusted uploader\n")
			}
		}
	}
	fmt.Println()
}
