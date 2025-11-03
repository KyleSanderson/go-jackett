# Implementation Summary: Torznab Specification Compliance

## Overview
This implementation brings the go-jackett library into full compliance with the Torznab v1.3 specification.

## Previous Issues Fixed

### 1. HTTP Connection Management
- **Problem**: Response bodies were not properly drained before closing, causing connection leaks
- **Solution**: Implemented `drainAndClose()` pattern that reads remaining data before closing
- **Impact**: Enables HTTP connection reuse, reduces resource consumption

### 2. Direct Tracker Support
- **Problem**: Library only supported Jackett proxy mode
- **Solution**: Added `DirectMode` configuration to query tracker APIs directly
- **Impact**: Better performance, no proxy overhead, works with trackers like MoreThanTV

## New Features Implemented (Torznab Spec v1.3)

### 1. Specialized Search Methods

#### TV Search (`t=tvsearch`)
- Full support for TV-specific parameters:
  - `tvdbid` - TheTVDB ID
  - `tvmazeid` - TVMaze ID
  - `rid` - TVRage ID (deprecated but supported)
  - `season` - Season number
  - `episode` - Episode number
  - `q` - Free text query
- Type-safe options struct: `TVSearchOptions`

#### Movie Search (`t=movie`)
- Full support for movie-specific parameters:
  - `imdbid` - IMDB ID (with or without tt prefix)
  - `tmdbid` - The Movie Database ID
  - `genre` - Genre filter
  - `year` - Release year
  - `q` - Free text query
- Type-safe options struct: `MovieSearchOptions`

#### Music Search (`t=music`)
- Full support for music-specific parameters:
  - `artist` - Artist name
  - `album` - Album name
  - `label` - Record label
  - `track` - Track name
  - `year` - Release year
  - `genre` - Genre filter
  - `q` - Free text query
- Type-safe options struct: `MusicSearchOptions`

#### Book Search (`t=book`)
- Full support for book-specific parameters:
  - `author` - Author name
  - `title` - Book title
  - `year` - Publication year
  - `genre` - Genre filter
  - `q` - Free text query
- Type-safe options struct: `BookSearchOptions`

### 2. Category Constants
Added all standard Torznab categories defined in the specification:
- TV categories (5000-5090)
- Movie categories (2000-2070)
- Audio/Music categories (3000-3060)
- Books categories (8000-8060)
- PC categories (4000-4070)
- XXX categories (6000-6090)
- Other categories (7000-7020)

### 3. Torznab Attribute Helpers
Created `TorznabItem` type with 30+ helper methods for easy access to extended attributes:

#### Torrent Health & Identifiers
- `Seeders()` - Number of seeders
- `Leechers()` - Number of leechers
- `Peers()` - Total peers
- `InfoHash()` - Torrent info hash
- `MagnetURL()` - Magnet link

#### Freeleech Support
- `IsFreeleech()` - Check if torrent is freeleech
- `DownloadVolumeFactor()` - Download multiplier (0.0 = freeleech)
- `UploadVolumeFactor()` - Upload multiplier
- `MinimumRatio()` - Required ratio
- `MinimumSeedTime()` - Required seed time in seconds

#### TV Show Attributes
- `TVDBID()` - TheTVDB ID
- `TVMazeID()` - TVMaze ID
- `Season()` - Season number
- `Episode()` - Episode number

#### Movie Attributes
- `IMDBID()` - IMDB identifier
- `TMDBID()` - TMDB identifier
- `Year()` - Release year

#### Media Quality
- `Resolution()` - Video resolution
- `Video()` - Video codec
- `Audio()` - Audio codec
- `Language()` - Language
- `Subtitles()` - Subtitle languages

#### Music Attributes
- `Artist()` - Artist name
- `Album()` - Album name

#### Book Attributes
- `Author()` - Author name
- `BookTitle()` - Book title

#### Tags & Categories
- `Categories()` - All category IDs
- `Tags()` - All tags
- `HasTag(tag)` - Check for specific tag

### 4. Common Query Parameters
All search methods support:
- `limit` - Number of results to return
- `offset` - Pagination offset
- `cat` - Category filter (comma-separated)
- `extended` - Request all extended attributes

### 5. Context Support
All methods have context-aware variants:
- `TVSearchCtx()`
- `MovieSearchCtx()`
- `MusicSearchCtx()`
- `BookSearchCtx()`
- Enables timeout and cancellation

## Files Added/Modified

### New Files
1. `constants.go` - Category constants and attribute name constants
2. `torznab.go` - TorznabItem type with helper methods
3. `examples/comprehensive_usage.go` - Complete usage examples
4. `examples/direct_tracker_usage.go` - Direct mode examples

### Modified Files
1. `methods.go` - Added 8 new search methods with options structs
2. `http.go` - Added `drainAndClose()` helper, updated `buildUrl()` for direct mode
3. `jackett.go` - Added `DirectMode` configuration option
4. `README.md` - Comprehensive documentation with examples

## Compliance Summary

### Implemented Features ✅
- ✅ Generic search (`t=search`)
- ✅ TV search (`t=tvsearch`)
- ✅ Movie search (`t=movie`)
- ✅ Music search (`t=music`)
- ✅ Book search (`t=book`)
- ✅ Capabilities query (`t=caps`)
- ✅ All standard query parameters (`cat`, `limit`, `offset`, `extended`)
- ✅ All extended attributes parsing
- ✅ Category constants
- ✅ Direct tracker mode
- ✅ Connection reuse (drain-and-close)
- ✅ Context support
- ✅ Retry logic with backoff

### Not Implemented (Newznab-specific features marked optional)
- ❌ `t=details` - Returns details about a specific item (newznab only)
- ❌ `t=getnfo` - Returns NFO file (newznab only)
- ❌ `t=get` - Returns NZB file (newznab specific)
- ❌ `t=cart-add/cart-del` - Shopping cart (newznab only)
- ❌ `t=comments` - Comment system (newznab only)
- ❌ `t=register` - User registration (newznab only)
- ❌ `t=user` - User information (newznab only)

These are all marked as "newznab" specific in the spec and not applicable to torrent indexers.

## Usage Examples

### TV Search
```go
results, err := client.TVSearch(jackett.TVSearchOptions{
    TVDBID:   "280619",
    Season:   "6",
    Episode:  "5",
    Category: jackett.CategoryTVHD,
})

items := results.ToTorznabItems()
for _, item := range items {
    fmt.Printf("S%02dE%02d - Seeders: %d\n", 
        item.Season(), item.Episode(), item.Seeders())
    if item.IsFreeleech() {
        fmt.Println("FREELEECH!")
    }
}
```

### Movie Search
```go
results, err := client.MovieSearch(jackett.MovieSearchOptions{
    IMDBID:   "tt0816692",
    Category: jackett.CategoryMoviesUHD,
})
```

### Direct Tracker Mode
```go
client := jackett.NewClient(jackett.Config{
    Host:       "https://tracker.example.com/api/torznab",
    APIKey:     "your-api-key",
    DirectMode: true,
})
```

## Backward Compatibility
All existing code continues to work:
- `GetTorrents()` - Still works for generic searches
- `GetIndexers()` - Still works for capabilities
- `SearchDirect()` - Generic direct search still available
- No breaking changes to existing APIs

## Testing Recommendations
1. Test TV search with real TVDB IDs
2. Test movie search with IMDB IDs
3. Verify freeleech detection works
4. Test direct mode with actual tracker
5. Verify all attribute helpers return correct values
6. Test context cancellation
7. Verify connection reuse with multiple requests

## References
- [Torznab Specification v1.3](https://torznab.github.io/spec-1.3-draft/torznab/Specification-v1.3.html)
- [Sonarr Torznab Implementation Guide](https://github.com/Sonarr/Sonarr/wiki/Implementing-a-Torznab-indexer)
