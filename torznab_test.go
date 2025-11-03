package jackett

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTorznabItemHelpers(t *testing.T) {
	// Create a test RSS with various attributes
	rss := Rss{
		Channel: struct {
			Text string `xml:",chardata"`
			Link struct {
				Text string `xml:",chardata"`
				Href string `xml:"href,attr"`
				Rel  string `xml:"rel,attr"`
				Type string `xml:"type,attr"`
			} `xml:"link"`
			Title       string `xml:"title"`
			Description string `xml:"description"`
			Language    string `xml:"language"`
			Category    string `xml:"category"`
			Item        []struct {
				Text           string `xml:",chardata"`
				Title          string `xml:"title"`
				Guid           string `xml:"guid"`
				Jackettindexer struct {
					Text string `xml:",chardata"`
					ID   string `xml:"id,attr"`
				} `xml:"jackettindexer"`
				Type        string   `xml:"type"`
				Comments    string   `xml:"comments"`
				PubDate     string   `xml:"pubDate"`
				Size        string   `xml:"size"`
				Files       string   `xml:"files"`
				Grabs       string   `xml:"grabs"`
				Description string   `xml:"description"`
				Link        string   `xml:"link"`
				Category    []string `xml:"category"`
				Enclosure   struct {
					Text   string `xml:",chardata"`
					URL    string `xml:"url,attr"`
					Length string `xml:"length,attr"`
					Type   string `xml:"type,attr"`
				} `xml:"enclosure"`
				Attr []struct {
					Text  string `xml:",chardata"`
					Name  string `xml:"name,attr"`
					Value string `xml:"value,attr"`
				} `xml:"attr"`
			} `xml:"item"`
		}{
			Item: []struct {
				Text           string `xml:",chardata"`
				Title          string `xml:"title"`
				Guid           string `xml:"guid"`
				Jackettindexer struct {
					Text string `xml:",chardata"`
					ID   string `xml:"id,attr"`
				} `xml:"jackettindexer"`
				Type        string   `xml:"type"`
				Comments    string   `xml:"comments"`
				PubDate     string   `xml:"pubDate"`
				Size        string   `xml:"size"`
				Files       string   `xml:"files"`
				Grabs       string   `xml:"grabs"`
				Description string   `xml:"description"`
				Link        string   `xml:"link"`
				Category    []string `xml:"category"`
				Enclosure   struct {
					Text   string `xml:",chardata"`
					URL    string `xml:"url,attr"`
					Length string `xml:"length,attr"`
					Type   string `xml:"type,attr"`
				} `xml:"enclosure"`
				Attr []struct {
					Text  string `xml:",chardata"`
					Name  string `xml:"name,attr"`
					Value string `xml:"value,attr"`
				} `xml:"attr"`
			}{
				{
					Title:       "The Expanse S06E05 1080p",
					Guid:        "test-guid-1",
					Link:        "http://example.com/download/1",
					PubDate:     "Mon, 01 Jan 2024 00:00:00 +0000",
					Size:        "1073741824",
					Files:       "10",
					Grabs:       "100",
					Description: "Test description",
					Category:    []string{"5000", "5040"},
					Enclosure: struct {
						Text   string `xml:",chardata"`
						URL    string `xml:"url,attr"`
						Length string `xml:"length,attr"`
						Type   string `xml:"type,attr"`
					}{
						URL:    "http://example.com/download/1.torrent",
						Length: "1073741824",
					},
					Attr: []struct {
						Text  string `xml:",chardata"`
						Name  string `xml:"name,attr"`
						Value string `xml:"value,attr"`
					}{
						{Name: "seeders", Value: "50"},
						{Name: "leechers", Value: "10"},
						{Name: "peers", Value: "60"},
						{Name: "infohash", Value: "abcdef1234567890"},
						{Name: "magneturl", Value: "magnet:?xt=urn:btih:abcdef1234567890"},
						{Name: "downloadvolumefactor", Value: "0"},
						{Name: "uploadvolumefactor", Value: "2"},
						{Name: "minimumratio", Value: "1.5"},
						{Name: "minimumseedtime", Value: "86400"},
						{Name: "tvdbid", Value: "280619"},
						{Name: "tvmazeid", Value: "12345"},
						{Name: "season", Value: "6"},
						{Name: "episode", Value: "5"},
						{Name: "imdbid", Value: "tt1234567"},
						{Name: "tmdbid", Value: "54321"},
						{Name: "genre", Value: "Sci-Fi"},
						{Name: "year", Value: "2021"},
						{Name: "resolution", Value: "1920x1080"},
						{Name: "video", Value: "x264"},
						{Name: "audio", Value: "AAC 5.1"},
						{Name: "language", Value: "English"},
						{Name: "subs", Value: "English, Spanish"},
						{Name: "artist", Value: "Test Artist"},
						{Name: "album", Value: "Test Album"},
						{Name: "author", Value: "Test Author"},
						{Name: "booktitle", Value: "Test Book"},
						{Name: "tag", Value: "internal"},
						{Name: "tag", Value: "trusted"},
						{Name: "category", Value: "5000"},
						{Name: "category", Value: "5040"},
					},
				},
			},
		},
	}

	items := rss.ToTorznabItems()
	assert.Len(t, items, 1)

	item := items[0]

	// Test basic fields
	assert.Equal(t, "The Expanse S06E05 1080p", item.Title)
	assert.Equal(t, "test-guid-1", item.Guid)
	assert.Equal(t, "http://example.com/download/1", item.Link)
	assert.Equal(t, "1073741824", item.Size)
	assert.Equal(t, "http://example.com/download/1.torrent", item.EnclosureURL)

	// Test seeders/leechers/peers
	assert.Equal(t, 50, item.Seeders())
	assert.Equal(t, 10, item.Leechers())
	assert.Equal(t, 60, item.Peers())

	// Test torrent identifiers
	assert.Equal(t, "abcdef1234567890", item.InfoHash())
	assert.Equal(t, "magnet:?xt=urn:btih:abcdef1234567890", item.MagnetURL())

	// Test freeleech
	assert.True(t, item.IsFreeleech())
	assert.Equal(t, 0.0, item.DownloadVolumeFactor())
	assert.Equal(t, 2.0, item.UploadVolumeFactor())

	// Test seeding requirements
	assert.Equal(t, 1.5, item.MinimumRatio())
	assert.Equal(t, int64(86400), item.MinimumSeedTime())

	// Test TV attributes
	assert.Equal(t, "280619", item.TVDBID())
	assert.Equal(t, "12345", item.TVMazeID())
	assert.Equal(t, 6, item.Season())
	assert.Equal(t, 5, item.Episode())

	// Test movie attributes
	assert.Equal(t, "tt1234567", item.IMDBID())
	assert.Equal(t, "54321", item.TMDBID())
	assert.Equal(t, 2021, item.Year())

	// Test media quality
	assert.Equal(t, "1920x1080", item.Resolution())
	assert.Equal(t, "x264", item.Video())
	assert.Equal(t, "AAC 5.1", item.Audio())
	assert.Equal(t, "English", item.Language())
	assert.Equal(t, "English, Spanish", item.Subtitles())

	// Test genre
	assert.Equal(t, "Sci-Fi", item.Genre())

	// Test music attributes
	assert.Equal(t, "Test Artist", item.Artist())
	assert.Equal(t, "Test Album", item.Album())

	// Test book attributes
	assert.Equal(t, "Test Author", item.Author())
	assert.Equal(t, "Test Book", item.BookTitle())

	// Test tags
	tags := item.Tags()
	assert.Len(t, tags, 2)
	assert.Contains(t, tags, "internal")
	assert.Contains(t, tags, "trusted")
	assert.True(t, item.HasTag("internal"))
	assert.True(t, item.HasTag("INTERNAL")) // Case insensitive
	assert.False(t, item.HasTag("nonexistent"))

	// Test categories
	cats := item.Categories()
	assert.Len(t, cats, 2)
	assert.Contains(t, cats, "5000")
	assert.Contains(t, cats, "5040")
}

func TestTorznabItemDefaults(t *testing.T) {
	// Test item with no attributes
	rss := Rss{
		Channel: struct {
			Text string `xml:",chardata"`
			Link struct {
				Text string `xml:",chardata"`
				Href string `xml:"href,attr"`
				Rel  string `xml:"rel,attr"`
				Type string `xml:"type,attr"`
			} `xml:"link"`
			Title       string `xml:"title"`
			Description string `xml:"description"`
			Language    string `xml:"language"`
			Category    string `xml:"category"`
			Item        []struct {
				Text           string `xml:",chardata"`
				Title          string `xml:"title"`
				Guid           string `xml:"guid"`
				Jackettindexer struct {
					Text string `xml:",chardata"`
					ID   string `xml:"id,attr"`
				} `xml:"jackettindexer"`
				Type        string   `xml:"type"`
				Comments    string   `xml:"comments"`
				PubDate     string   `xml:"pubDate"`
				Size        string   `xml:"size"`
				Files       string   `xml:"files"`
				Grabs       string   `xml:"grabs"`
				Description string   `xml:"description"`
				Link        string   `xml:"link"`
				Category    []string `xml:"category"`
				Enclosure   struct {
					Text   string `xml:",chardata"`
					URL    string `xml:"url,attr"`
					Length string `xml:"length,attr"`
					Type   string `xml:"type,attr"`
				} `xml:"enclosure"`
				Attr []struct {
					Text  string `xml:",chardata"`
					Name  string `xml:"name,attr"`
					Value string `xml:"value,attr"`
				} `xml:"attr"`
			} `xml:"item"`
		}{
			Item: []struct {
				Text           string `xml:",chardata"`
				Title          string `xml:"title"`
				Guid           string `xml:"guid"`
				Jackettindexer struct {
					Text string `xml:",chardata"`
					ID   string `xml:"id,attr"`
				} `xml:"jackettindexer"`
				Type        string   `xml:"type"`
				Comments    string   `xml:"comments"`
				PubDate     string   `xml:"pubDate"`
				Size        string   `xml:"size"`
				Files       string   `xml:"files"`
				Grabs       string   `xml:"grabs"`
				Description string   `xml:"description"`
				Link        string   `xml:"link"`
				Category    []string `xml:"category"`
				Enclosure   struct {
					Text   string `xml:",chardata"`
					URL    string `xml:"url,attr"`
					Length string `xml:"length,attr"`
					Type   string `xml:"type,attr"`
				} `xml:"enclosure"`
				Attr []struct {
					Text  string `xml:",chardata"`
					Name  string `xml:"name,attr"`
					Value string `xml:"value,attr"`
				} `xml:"attr"`
			}{
				{
					Title:    "Test Item",
					Category: []string{"5000"},
				},
			},
		},
	}

	items := rss.ToTorznabItems()
	assert.Len(t, items, 1)

	item := items[0]

	// Test defaults
	assert.Equal(t, 0, item.Seeders())
	assert.Equal(t, 0, item.Leechers())
	assert.Equal(t, 0, item.Peers())
	assert.Equal(t, "", item.InfoHash())
	assert.Equal(t, "", item.MagnetURL())
	assert.False(t, item.IsFreeleech())
	assert.Equal(t, 1.0, item.DownloadVolumeFactor()) // Default to 1.0
	assert.Equal(t, 1.0, item.UploadVolumeFactor())   // Default to 1.0
	assert.Equal(t, 0.0, item.MinimumRatio())
	assert.Equal(t, int64(0), item.MinimumSeedTime())
	assert.Equal(t, "", item.TVDBID())
	assert.Equal(t, 0, item.Season())
	assert.Equal(t, 0, item.Episode())
	assert.Equal(t, "", item.IMDBID())
	assert.Equal(t, 0, item.Year())
	assert.Len(t, item.Tags(), 0)

	// Categories should fall back to RSS category
	cats := item.Categories()
	assert.Len(t, cats, 1)
	assert.Contains(t, cats, "5000")
}

func TestTorznabItemPeersCalculation(t *testing.T) {
	// Test when peers attribute is not present, it calculates from seeders + leechers
	rss := Rss{
		Channel: struct {
			Text string `xml:",chardata"`
			Link struct {
				Text string `xml:",chardata"`
				Href string `xml:"href,attr"`
				Rel  string `xml:"rel,attr"`
				Type string `xml:"type,attr"`
			} `xml:"link"`
			Title       string `xml:"title"`
			Description string `xml:"description"`
			Language    string `xml:"language"`
			Category    string `xml:"category"`
			Item        []struct {
				Text           string `xml:",chardata"`
				Title          string `xml:"title"`
				Guid           string `xml:"guid"`
				Jackettindexer struct {
					Text string `xml:",chardata"`
					ID   string `xml:"id,attr"`
				} `xml:"jackettindexer"`
				Type        string   `xml:"type"`
				Comments    string   `xml:"comments"`
				PubDate     string   `xml:"pubDate"`
				Size        string   `xml:"size"`
				Files       string   `xml:"files"`
				Grabs       string   `xml:"grabs"`
				Description string   `xml:"description"`
				Link        string   `xml:"link"`
				Category    []string `xml:"category"`
				Enclosure   struct {
					Text   string `xml:",chardata"`
					URL    string `xml:"url,attr"`
					Length string `xml:"length,attr"`
					Type   string `xml:"type,attr"`
				} `xml:"enclosure"`
				Attr []struct {
					Text  string `xml:",chardata"`
					Name  string `xml:"name,attr"`
					Value string `xml:"value,attr"`
				} `xml:"attr"`
			} `xml:"item"`
		}{
			Item: []struct {
				Text           string `xml:",chardata"`
				Title          string `xml:"title"`
				Guid           string `xml:"guid"`
				Jackettindexer struct {
					Text string `xml:",chardata"`
					ID   string `xml:"id,attr"`
				} `xml:"jackettindexer"`
				Type        string   `xml:"type"`
				Comments    string   `xml:"comments"`
				PubDate     string   `xml:"pubDate"`
				Size        string   `xml:"size"`
				Files       string   `xml:"files"`
				Grabs       string   `xml:"grabs"`
				Description string   `xml:"description"`
				Link        string   `xml:"link"`
				Category    []string `xml:"category"`
				Enclosure   struct {
					Text   string `xml:",chardata"`
					URL    string `xml:"url,attr"`
					Length string `xml:"length,attr"`
					Type   string `xml:"type,attr"`
				} `xml:"enclosure"`
				Attr []struct {
					Text  string `xml:",chardata"`
					Name  string `xml:"name,attr"`
					Value string `xml:"value,attr"`
				} `xml:"attr"`
			}{
				{
					Title: "Test",
					Attr: []struct {
						Text  string `xml:",chardata"`
						Name  string `xml:"name,attr"`
						Value string `xml:"value,attr"`
					}{
						{Name: "seeders", Value: "30"},
						{Name: "leechers", Value: "20"},
					},
				},
			},
		},
	}

	items := rss.ToTorznabItems()
	item := items[0]

	// Should calculate peers from seeders + leechers
	assert.Equal(t, 50, item.Peers())
}

func TestTorznabItemIMDBFallback(t *testing.T) {
	// Test IMDB attribute fallback (imdbid vs imdb)
	rss := Rss{
		Channel: struct {
			Text string `xml:",chardata"`
			Link struct {
				Text string `xml:",chardata"`
				Href string `xml:"href,attr"`
				Rel  string `xml:"rel,attr"`
				Type string `xml:"type,attr"`
			} `xml:"link"`
			Title       string `xml:"title"`
			Description string `xml:"description"`
			Language    string `xml:"language"`
			Category    string `xml:"category"`
			Item        []struct {
				Text           string `xml:",chardata"`
				Title          string `xml:"title"`
				Guid           string `xml:"guid"`
				Jackettindexer struct {
					Text string `xml:",chardata"`
					ID   string `xml:"id,attr"`
				} `xml:"jackettindexer"`
				Type        string   `xml:"type"`
				Comments    string   `xml:"comments"`
				PubDate     string   `xml:"pubDate"`
				Size        string   `xml:"size"`
				Files       string   `xml:"files"`
				Grabs       string   `xml:"grabs"`
				Description string   `xml:"description"`
				Link        string   `xml:"link"`
				Category    []string `xml:"category"`
				Enclosure   struct {
					Text   string `xml:",chardata"`
					URL    string `xml:"url,attr"`
					Length string `xml:"length,attr"`
					Type   string `xml:"type,attr"`
				} `xml:"enclosure"`
				Attr []struct {
					Text  string `xml:",chardata"`
					Name  string `xml:"name,attr"`
					Value string `xml:"value,attr"`
				} `xml:"attr"`
			} `xml:"item"`
		}{
			Item: []struct {
				Text           string `xml:",chardata"`
				Title          string `xml:"title"`
				Guid           string `xml:"guid"`
				Jackettindexer struct {
					Text string `xml:",chardata"`
					ID   string `xml:"id,attr"`
				} `xml:"jackettindexer"`
				Type        string   `xml:"type"`
				Comments    string   `xml:"comments"`
				PubDate     string   `xml:"pubDate"`
				Size        string   `xml:"size"`
				Files       string   `xml:"files"`
				Grabs       string   `xml:"grabs"`
				Description string   `xml:"description"`
				Link        string   `xml:"link"`
				Category    []string `xml:"category"`
				Enclosure   struct {
					Text   string `xml:",chardata"`
					URL    string `xml:"url,attr"`
					Length string `xml:"length,attr"`
					Type   string `xml:"type,attr"`
				} `xml:"enclosure"`
				Attr []struct {
					Text  string `xml:",chardata"`
					Name  string `xml:"name,attr"`
					Value string `xml:"value,attr"`
				} `xml:"attr"`
			}{
				{
					Title: "Test",
					Attr: []struct {
						Text  string `xml:",chardata"`
						Name  string `xml:"name,attr"`
						Value string `xml:"value,attr"`
					}{
						{Name: "imdb", Value: "tt9876543"},
					},
				},
			},
		},
	}

	items := rss.ToTorznabItems()
	item := items[0]

	// Should use "imdb" attribute as fallback
	assert.Equal(t, "tt9876543", item.IMDBID())
}

func TestGetAttrValues(t *testing.T) {
	item := TorznabItem{
		Attributes: map[string][]string{
			"category": {"5000", "5040", "5070"},
			"tag":      {"internal", "trusted"},
		},
	}

	categories := item.GetAttrValues("category")
	assert.Len(t, categories, 3)
	assert.Contains(t, categories, "5000")
	assert.Contains(t, categories, "5040")
	assert.Contains(t, categories, "5070")

	tags := item.GetAttrValues("tag")
	assert.Len(t, tags, 2)

	// Non-existent attribute
	nonExistent := item.GetAttrValues("nonexistent")
	assert.Len(t, nonExistent, 0)
}

func TestGetAttrMethods(t *testing.T) {
	item := TorznabItem{
		Attributes: map[string][]string{
			"teststring": {"value1"},
			"testint":    {"42"},
			"testint64":  {"9999999999"},
			"testfloat":  {"3.14"},
			"invalid":    {"not-a-number"},
		},
	}

	// Test GetAttr
	assert.Equal(t, "value1", item.GetAttr("teststring"))
	assert.Equal(t, "", item.GetAttr("nonexistent"))

	// Test GetAttrInt
	assert.Equal(t, 42, item.GetAttrInt("testint"))
	assert.Equal(t, 0, item.GetAttrInt("nonexistent"))
	assert.Equal(t, 0, item.GetAttrInt("invalid"))

	// Test GetAttrInt64
	assert.Equal(t, int64(9999999999), item.GetAttrInt64("testint64"))
	assert.Equal(t, int64(0), item.GetAttrInt64("nonexistent"))

	// Test GetAttrFloat
	assert.Equal(t, 3.14, item.GetAttrFloat("testfloat"))
	assert.Equal(t, 0.0, item.GetAttrFloat("nonexistent"))
	assert.Equal(t, 0.0, item.GetAttrFloat("invalid"))
}
