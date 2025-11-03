# go-jackett

Go library for communicating with Jackett and torznab APIs.

## Features

- Proxy searches through Jackett to multiple indexers
- Direct communication with tracker torznab APIs (bypassing Jackett)
- Proper HTTP connection reuse with drain-and-close pattern
- Context support for timeout and cancellation
- Retry logic with exponential backoff

## Installation

```bash
go get github.com/autobrr/go-jackett
```

## Usage

### Jackett Proxy Mode (Default)

Use Jackett as a proxy to search multiple indexers:

```go
package main

import (
    "fmt"
    "log"
    "github.com/autobrr/go-jackett"
)

func main() {
    client := jackett.NewClient(jackett.Config{
        Host:   "http://localhost:9117",
        APIKey: "your-jackett-api-key",
    })

    // Search all indexers via Jackett
    opts := map[string]string{
        "t": "search",
        "q": "ubuntu",
    }
    
    results, err := client.GetTorrents("all", opts)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Found %d results\n", len(results.Channel.Items))
}
```

### Direct Tracker Mode

Connect directly to a tracker's torznab API (e.g., MoreThanTV, RED, etc.):

```go
package main

import (
    "fmt"
    "log"
    "github.com/autobrr/go-jackett"
)

func main() {
    client := jackett.NewClient(jackett.Config{
        Host:       "https://www.morethantv.me/api/torznab",
        APIKey:     "your-tracker-api-key",
        DirectMode: true, // Enable direct mode
        Timeout:    30,
    })

    // Search directly on the tracker
    results, err := client.SearchDirect("ubuntu", map[string]string{
        "limit": "50",
        "cat":   "5000", // TV category
    })
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Found %d results\n", len(results.Channel.Items))
    
    // Get tracker capabilities
    caps, err := client.GetCapsDirect()
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Tracker: %s\n", caps.Indexer[0].Title)
}
```

### Using Context

All methods have context-aware variants for timeout and cancellation:

```go
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

results, err := client.SearchDirectCtx(ctx, "debian", nil)
if err != nil {
    log.Fatal(err)
}
```

## Configuration Options

| Option | Type | Description |
|--------|------|-------------|
| `Host` | string | Jackett URL or direct tracker torznab URL |
| `APIKey` | string | API key for authentication |
| `DirectMode` | bool | Enable direct tracker mode (bypass Jackett) |
| `TLSSkipVerify` | bool | Skip TLS certificate verification |
| `BasicUser` | string | HTTP Basic auth username |
| `BasicPass` | string | HTTP Basic auth password |
| `Timeout` | int | Request timeout in seconds (default: 60) |
| `Log` | *log.Logger | Custom logger |

## Specialized Search Methods

The library now supports all torznab search types as defined in the [torznab specification](https://torznab.github.io/spec-1.3-draft/):

### TV Search

Search for TV shows with specific parameters:

```go
results, err := client.TVSearch(jackett.TVSearchOptions{
    Query:    "The Expanse",
    TVDBID:   "280619",
    Season:   "6",
    Episode:  "5",
    Category: jackett.CategoryTVHD,
})

// Process results with helper methods
items := results.ToTorznabItems()
for _, item := range items {
    fmt.Printf("Title: %s\n", item.Title)
    fmt.Printf("Seeders: %d, Leechers: %d\n", item.Seeders(), item.Leechers())
    fmt.Printf("Freeleech: %v\n", item.IsFreeleech())
    fmt.Printf("InfoHash: %s\n", item.InfoHash())
}
```

**TVSearchOptions fields:**
- `Query` - Free text search
- `TVDBID` - TheTVDB ID
- `TVMazeID` - TVMaze ID
- `RageID` - TVRage ID (deprecated)
- `Season` - Season number
- `Episode` - Episode number
- `Category` - Category filter (use constants like `CategoryTVHD`)
- `Limit`, `Offset`, `Extended`

### Movie Search

Search for movies with IMDB/TMDB support:

```go
results, err := client.MovieSearch(jackett.MovieSearchOptions{
    IMDBID:   "tt0816692",  // Interstellar
    Year:     "2014",
    Category: jackett.CategoryMoviesHD,
})

items := results.ToTorznabItems()
for _, item := range items {
    fmt.Printf("%s - %s\n", item.Title, item.Resolution())
    fmt.Printf("IMDB: %s, Year: %d\n", item.IMDBID(), item.Year())
}
```

**MovieSearchOptions fields:**
- `Query` - Free text search
- `IMDBID` - IMDB ID (with or without 'tt' prefix)
- `TMDBID` - The Movie Database ID
- `Genre` - Genre filter
- `Year` - Release year
- `Category`, `Limit`, `Offset`, `Extended`

### Music Search

Search for music with artist/album support:

```go
results, err := client.MusicSearch(jackett.MusicSearchOptions{
    Artist:   "Pink Floyd",
    Album:    "Dark Side of the Moon",
    Year:     "1973",
    Category: jackett.CategoryAudioLossless,
})

items := results.ToTorznabItems()
for _, item := range items {
    fmt.Printf("%s by %s\n", item.Album(), item.Artist())
}
```

**MusicSearchOptions fields:**
- `Query` - Free text search
- `Artist` - Artist name
- `Album` - Album name
- `Label` - Record label
- `Track` - Track name
- `Year` - Release year
- `Genre` - Genre filter
- `Category`, `Limit`, `Offset`, `Extended`

### Book Search

Search for books with author/title support:

```go
results, err := client.BookSearch(jackett.BookSearchOptions{
    Author: "Isaac Asimov",
    Title:  "Foundation",
    Category: jackett.CategoryBooksEBook,
})

items := results.ToTorznabItems()
for _, item := range items {
    fmt.Printf("%s by %s\n", item.BookTitle(), item.Author())
}
```

**BookSearchOptions fields:**
- `Query` - Free text search
- `Author` - Author name
- `Title` - Book title
- `Year` - Publication year
- `Category`, `Limit`, `Offset`, `Extended`

## Torznab Attribute Helpers

The `TorznabItem` type provides convenient methods to access torznab extended attributes:

```go
items := results.ToTorznabItems()
for _, item := range items {
    // Torrent health
    fmt.Printf("Seeders: %d\n", item.Seeders())
    fmt.Printf("Leechers: %d\n", item.Leechers())
    fmt.Printf("Peers: %d\n", item.Peers())
    
    // Torrent identifiers
    fmt.Printf("InfoHash: %s\n", item.InfoHash())
    fmt.Printf("Magnet: %s\n", item.MagnetURL())
    
    // Freeleech detection
    if item.IsFreeleech() {
        fmt.Println("This is freeleech!")
    }
    fmt.Printf("Download factor: %.1f\n", item.DownloadVolumeFactor())
    fmt.Printf("Upload factor: %.1f\n", item.UploadVolumeFactor())
    
    // TV show attributes
    fmt.Printf("TVDB ID: %s\n", item.TVDBID())
    fmt.Printf("Season %d Episode %d\n", item.Season(), item.Episode())
    
    // Movie attributes
    fmt.Printf("IMDB: %s\n", item.IMDBID())
    fmt.Printf("TMDB: %s\n", item.TMDBID())
    
    // Media quality
    fmt.Printf("Resolution: %s\n", item.Resolution())
    fmt.Printf("Video: %s\n", item.Video())
    fmt.Printf("Audio: %s\n", item.Audio())
    
    // Categories and tags
    fmt.Printf("Categories: %v\n", item.Categories())
    fmt.Printf("Tags: %v\n", item.Tags())
    if item.HasTag("internal") {
        fmt.Println("Internal release")
    }
}
```

## Category Constants

The library provides constants for all standard torznab categories:

```go
// TV
jackett.CategoryTV, jackett.CategoryTVHD, jackett.CategoryTVSD, 
jackett.CategoryTVUHD, jackett.CategoryTVAnime

// Movies
jackett.CategoryMovies, jackett.CategoryMoviesHD, jackett.CategoryMoviesSD,
jackett.CategoryMoviesUHD, jackett.CategoryMoviesBluRay

// Music
jackett.CategoryAudio, jackett.CategoryAudioMP3, jackett.CategoryAudioLossless,
jackett.CategoryAudioAudiobook

// Books
jackett.CategoryBooks, jackett.CategoryBooksEBook, jackett.CategoryBooksComics

// And many more - see constants.go
```

## Methods

### Jackett Proxy Methods
- `GetIndexers()` / `GetIndexersCtx()` - Get configured indexers
- `GetTorrents(indexer, opts)` / `GetTorrentsCtx()` - Search torrents via Jackett
- `GetEnclosure(url)` / `GetEnclosureCtx()` - Download torrent file

### Direct Tracker Methods
- `SearchDirect(query, opts)` / `SearchDirectCtx()` - Generic search on tracker
- `GetCapsDirect()` / `GetCapsDirectCtx()` - Get tracker capabilities

### Specialized Search Methods (work in both modes)
- `TVSearch(opts)` / `TVSearchCtx()` - TV show search with TVDB/TVMaze support
- `MovieSearch(opts)` / `MovieSearchCtx()` - Movie search with IMDB/TMDB support
- `MusicSearch(opts)` / `MusicSearchCtx()` - Music search with artist/album support
- `BookSearch(opts)` / `BookSearchCtx()` - Book search with author/title support

## Improvements

This library implements best practices for HTTP client usage:

1. **Drain-and-Close Pattern**: All HTTP response bodies are properly drained before closing to ensure connection reuse and prevent resource leaks
2. **Direct Tracker Support**: Bypass Jackett and query tracker torznab APIs directly for better performance and reliability
3. **Context Support**: All methods support context for proper timeout and cancellation handling

## License

See LICENSE file.
