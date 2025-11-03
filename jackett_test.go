package jackett

import (
	"context"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name   string
		config Config
		check  func(t *testing.T, c *Client)
	}{
		{
			name: "default configuration",
			config: Config{
				Host:   "http://localhost:9117",
				APIKey: "test-key",
			},
			check: func(t *testing.T, c *Client) {
				assert.NotNil(t, c)
				assert.NotNil(t, c.http)
				assert.Equal(t, DefaultTimeout, c.timeout)
				assert.NotNil(t, c.log)
			},
		},
		{
			name: "with custom timeout",
			config: Config{
				Host:    "http://localhost:9117",
				APIKey:  "test-key",
				Timeout: 30,
			},
			check: func(t *testing.T, c *Client) {
				assert.Equal(t, 30*time.Second, c.timeout)
			},
		},
		{
			name: "with custom logger",
			config: Config{
				Host:   "http://localhost:9117",
				APIKey: "test-key",
				Log:    log.New(io.Discard, "test: ", log.LstdFlags),
			},
			check: func(t *testing.T, c *Client) {
				assert.NotNil(t, c.log)
			},
		},
		{
			name: "with direct mode",
			config: Config{
				Host:       "https://tracker.example.com/api/torznab",
				APIKey:     "test-key",
				DirectMode: true,
			},
			check: func(t *testing.T, c *Client) {
				assert.True(t, c.cfg.DirectMode)
			},
		},
		{
			name: "with basic auth",
			config: Config{
				Host:      "http://localhost:9117",
				APIKey:    "test-key",
				BasicUser: "user",
				BasicPass: "pass",
			},
			check: func(t *testing.T, c *Client) {
				assert.Equal(t, "user", c.cfg.BasicUser)
				assert.Equal(t, "pass", c.cfg.BasicPass)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient(tt.config)
			tt.check(t, client)
		})
	}
}

func TestGetIndexers(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v2.0/indexers/all/results/torznab/api", r.URL.Path)
		assert.Equal(t, "indexers", r.URL.Query().Get("t"))
		assert.Equal(t, "true", r.URL.Query().Get("configured"))
		assert.Equal(t, "test-key", r.URL.Query().Get("apikey"))

		w.Header().Set("Content-Type", "application/xml")
		response := `<?xml version="1.0" encoding="UTF-8"?>
<indexers>
	<indexer id="test-indexer" configured="true">
		<title>Test Indexer</title>
		<description>Test Description</description>
	</indexer>
</indexers>`
		w.Write([]byte(response))
	}))
	defer server.Close()

	client := NewClient(Config{
		Host:   server.URL,
		APIKey: "test-key",
	})

	indexers, err := client.GetIndexers()
	require.NoError(t, err)
	assert.Len(t, indexers.Indexer, 1)
	assert.Equal(t, "test-indexer", indexers.Indexer[0].ID)
	assert.Equal(t, "Test Indexer", indexers.Indexer[0].Title)
}

func TestGetIndexersCtx(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		response := `<?xml version="1.0" encoding="UTF-8"?>
<indexers>
	<indexer id="test-indexer" configured="true">
		<title>Test Indexer</title>
	</indexer>
</indexers>`
		w.Write([]byte(response))
	}))
	defer server.Close()

	client := NewClient(Config{
		Host:   server.URL,
		APIKey: "test-key",
	})

	ctx := context.Background()
	indexers, err := client.GetIndexersCtx(ctx)
	require.NoError(t, err)
	assert.Len(t, indexers.Indexer, 1)
}

func TestGetIndexersError(t *testing.T) {
	client := NewClient(Config{
		Host:   "http://invalid-host-that-does-not-exist-12345.com",
		APIKey: "test-key",
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err := client.GetIndexersCtx(ctx)
	assert.Error(t, err)
}

func TestGetTorrents(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.Path, "/api/v2.0/indexers/")
		assert.Contains(t, r.URL.Path, "/results/torznab/api")
		assert.Equal(t, "search", r.URL.Query().Get("t"))
		assert.Equal(t, "ubuntu", r.URL.Query().Get("q"))

		w.Header().Set("Content-Type", "application/xml")
		response := `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
	<channel>
		<title>Test Results</title>
		<item>
			<title>Ubuntu 22.04 LTS</title>
			<guid>test-guid-1</guid>
			<link>http://example.com/download/1</link>
			<pubDate>Mon, 01 Jan 2024 00:00:00 +0000</pubDate>
		</item>
	</channel>
</rss>`
		w.Write([]byte(response))
	}))
	defer server.Close()

	client := NewClient(Config{
		Host:   server.URL,
		APIKey: "test-key",
	})

	opts := map[string]string{
		"t": "search",
		"q": "ubuntu",
	}

	results, err := client.GetTorrents("all", opts)
	require.NoError(t, err)
	assert.Len(t, results.Channel.Item, 1)
	assert.Equal(t, "Ubuntu 22.04 LTS", results.Channel.Item[0].Title)
}

func TestGetTorrentsCtx(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		response := `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
	<channel>
		<item>
			<title>Test Item</title>
		</item>
	</channel>
</rss>`
		w.Write([]byte(response))
	}))
	defer server.Close()

	client := NewClient(Config{
		Host:   server.URL,
		APIKey: "test-key",
	})

	ctx := context.Background()
	results, err := client.GetTorrentsCtx(ctx, "all", map[string]string{"t": "search"})
	require.NoError(t, err)
	assert.Len(t, results.Channel.Item, 1)
}

func TestGetEnclosure(t *testing.T) {
	expectedContent := []byte("torrent file content")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/x-bittorrent")
		w.Write(expectedContent)
	}))
	defer server.Close()

	client := NewClient(Config{
		Host:   server.URL,
		APIKey: "test-key",
	})

	content, err := client.GetEnclosure(server.URL + "/download/test.torrent")
	require.NoError(t, err)
	assert.Equal(t, expectedContent, content)
}

func TestGetEnclosureCtx(t *testing.T) {
	expectedContent := []byte("torrent file content")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(expectedContent)
	}))
	defer server.Close()

	client := NewClient(Config{
		Host: server.URL,
	})

	ctx := context.Background()
	content, err := client.GetEnclosureCtx(ctx, server.URL+"/download")
	require.NoError(t, err)
	assert.Equal(t, expectedContent, content)
}

func TestSearchDirect(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "search", r.URL.Query().Get("t"))
		assert.Equal(t, "ubuntu", r.URL.Query().Get("q"))
		assert.Equal(t, "test-key", r.URL.Query().Get("apikey"))

		w.Header().Set("Content-Type", "application/xml")
		response := `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
	<channel>
		<item>
			<title>Ubuntu Direct</title>
		</item>
	</channel>
</rss>`
		w.Write([]byte(response))
	}))
	defer server.Close()

	client := NewClient(Config{
		Host:       server.URL,
		APIKey:     "test-key",
		DirectMode: true,
	})

	results, err := client.SearchDirect("ubuntu", nil)
	require.NoError(t, err)
	assert.Len(t, results.Channel.Item, 1)
	assert.Equal(t, "Ubuntu Direct", results.Channel.Item[0].Title)
}

func TestSearchDirectCtx(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		response := `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
	<channel>
		<item>
			<title>Test</title>
		</item>
	</channel>
</rss>`
		w.Write([]byte(response))
	}))
	defer server.Close()

	client := NewClient(Config{
		Host:       server.URL,
		APIKey:     "test-key",
		DirectMode: true,
	})

	ctx := context.Background()
	results, err := client.SearchDirectCtx(ctx, "test", map[string]string{"limit": "10"})
	require.NoError(t, err)
	assert.Len(t, results.Channel.Item, 1)
}

func TestGetCapsDirect(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "caps", r.URL.Query().Get("t"))

		w.Header().Set("Content-Type", "application/xml")
		response := `<?xml version="1.0" encoding="UTF-8"?>
<indexers>
	<indexer id="test-tracker">
		<title>Test Tracker</title>
	</indexer>
</indexers>`
		w.Write([]byte(response))
	}))
	defer server.Close()

	client := NewClient(Config{
		Host:       server.URL,
		APIKey:     "test-key",
		DirectMode: true,
	})

	caps, err := client.GetCapsDirect()
	require.NoError(t, err)
	assert.Len(t, caps.Indexer, 1)
	assert.Equal(t, "Test Tracker", caps.Indexer[0].Title)
}

func TestGetCapsDirectCtx(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		response := `<?xml version="1.0" encoding="UTF-8"?>
<indexers>
	<indexer id="test">
		<title>Test</title>
	</indexer>
</indexers>`
		w.Write([]byte(response))
	}))
	defer server.Close()

	client := NewClient(Config{
		Host:       server.URL,
		DirectMode: true,
	})

	ctx := context.Background()
	caps, err := client.GetCapsDirectCtx(ctx)
	require.NoError(t, err)
	assert.Len(t, caps.Indexer, 1)
}

func TestTVSearch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "tvsearch", r.URL.Query().Get("t"))
		assert.Equal(t, "The Expanse", r.URL.Query().Get("q"))
		assert.Equal(t, "280619", r.URL.Query().Get("tvdbid"))
		assert.Equal(t, "12345", r.URL.Query().Get("tvmazeid"))
		assert.Equal(t, "6", r.URL.Query().Get("season"))
		assert.Equal(t, "5", r.URL.Query().Get("ep"))

		w.Header().Set("Content-Type", "application/xml")
		response := `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
	<channel>
		<item>
			<title>The Expanse S06E05</title>
		</item>
	</channel>
</rss>`
		w.Write([]byte(response))
	}))
	defer server.Close()

	client := NewClient(Config{
		Host:   server.URL,
		APIKey: "test-key",
	})

	results, err := client.TVSearch(TVSearchOptions{
		Query:    "The Expanse",
		TVDBID:   "280619",
		TVMazeID: "12345",
		Season:   "6",
		Episode:  "5",
	})
	require.NoError(t, err)
	assert.Len(t, results.Channel.Item, 1)
}

func TestTVSearchCtx(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		response := `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
	<channel>
		<item>
			<title>TV Show</title>
		</item>
	</channel>
</rss>`
		w.Write([]byte(response))
	}))
	defer server.Close()

	client := NewClient(Config{
		Host:       server.URL,
		DirectMode: true,
	})

	ctx := context.Background()
	results, err := client.TVSearchCtx(ctx, TVSearchOptions{
		Query:    "test",
		RageID:   "12345",
		IMDBID:   "tt1234567",
		Limit:    "50",
		Offset:   "10",
		Category: CategoryTVHD,
		Extended: "1",
	})
	require.NoError(t, err)
	assert.Len(t, results.Channel.Item, 1)
}

func TestMovieSearch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "movie", r.URL.Query().Get("t"))
		assert.Equal(t, "tt0816692", r.URL.Query().Get("imdbid"))
		assert.Equal(t, "2014", r.URL.Query().Get("year"))

		w.Header().Set("Content-Type", "application/xml")
		response := `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
	<channel>
		<item>
			<title>Interstellar (2014)</title>
		</item>
	</channel>
</rss>`
		w.Write([]byte(response))
	}))
	defer server.Close()

	client := NewClient(Config{
		Host:   server.URL,
		APIKey: "test-key",
	})

	results, err := client.MovieSearch(MovieSearchOptions{
		IMDBID: "tt0816692",
		Year:   "2014",
	})
	require.NoError(t, err)
	assert.Len(t, results.Channel.Item, 1)
}

func TestMovieSearchCtx(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		response := `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
	<channel>
		<item>
			<title>Movie</title>
		</item>
	</channel>
</rss>`
		w.Write([]byte(response))
	}))
	defer server.Close()

	client := NewClient(Config{
		Host:       server.URL,
		DirectMode: true,
	})

	ctx := context.Background()
	results, err := client.MovieSearchCtx(ctx, MovieSearchOptions{
		Query:    "test",
		TMDBID:   "12345",
		Genre:    "sci-fi",
		Limit:    "25",
		Offset:   "0",
		Category: CategoryMoviesHD,
		Extended: "1",
	})
	require.NoError(t, err)
	assert.Len(t, results.Channel.Item, 1)
}

func TestMusicSearch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "music", r.URL.Query().Get("t"))
		assert.Equal(t, "Pink Floyd", r.URL.Query().Get("artist"))
		assert.Equal(t, "Dark Side of the Moon", r.URL.Query().Get("album"))

		w.Header().Set("Content-Type", "application/xml")
		response := `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
	<channel>
		<item>
			<title>Pink Floyd - Dark Side of the Moon</title>
		</item>
	</channel>
</rss>`
		w.Write([]byte(response))
	}))
	defer server.Close()

	client := NewClient(Config{
		Host:   server.URL,
		APIKey: "test-key",
	})

	results, err := client.MusicSearch(MusicSearchOptions{
		Artist: "Pink Floyd",
		Album:  "Dark Side of the Moon",
	})
	require.NoError(t, err)
	assert.Len(t, results.Channel.Item, 1)
}

func TestMusicSearchCtx(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		response := `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
	<channel>
		<item>
			<title>Music</title>
		</item>
	</channel>
</rss>`
		w.Write([]byte(response))
	}))
	defer server.Close()

	client := NewClient(Config{
		Host:       server.URL,
		DirectMode: true,
	})

	ctx := context.Background()
	results, err := client.MusicSearchCtx(ctx, MusicSearchOptions{
		Query:    "test",
		Label:    "Columbia",
		Track:    "Track 1",
		Year:     "1973",
		Genre:    "rock",
		Limit:    "20",
		Offset:   "0",
		Category: CategoryAudioLossless,
		Extended: "1",
	})
	require.NoError(t, err)
	assert.Len(t, results.Channel.Item, 1)
}

func TestBookSearch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "book", r.URL.Query().Get("t"))
		assert.Equal(t, "Isaac Asimov", r.URL.Query().Get("author"))
		assert.Equal(t, "Foundation", r.URL.Query().Get("title"))

		w.Header().Set("Content-Type", "application/xml")
		response := `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
	<channel>
		<item>
			<title>Foundation by Isaac Asimov</title>
		</item>
	</channel>
</rss>`
		w.Write([]byte(response))
	}))
	defer server.Close()

	client := NewClient(Config{
		Host:   server.URL,
		APIKey: "test-key",
	})

	results, err := client.BookSearch(BookSearchOptions{
		Author: "Isaac Asimov",
		Title:  "Foundation",
	})
	require.NoError(t, err)
	assert.Len(t, results.Channel.Item, 1)
}

func TestBookSearchCtx(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		response := `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
	<channel>
		<item>
			<title>Book</title>
		</item>
	</channel>
</rss>`
		w.Write([]byte(response))
	}))
	defer server.Close()

	client := NewClient(Config{
		Host:       server.URL,
		DirectMode: true,
	})

	ctx := context.Background()
	results, err := client.BookSearchCtx(ctx, BookSearchOptions{
		Query:    "test",
		Year:     "2020",
		Genre:    "sci-fi",
		Limit:    "10",
		Offset:   "0",
		Category: CategoryBooksEBook,
		Extended: "1",
	})
	require.NoError(t, err)
	assert.Len(t, results.Channel.Item, 1)
}

func TestBuildURL(t *testing.T) {
	tests := []struct {
		name          string
		directMode    bool
		host          string
		endpoint      string
		params        map[string]string
		expectedPath  string
		expectedQuery map[string]string
	}{
		{
			name:          "jackett proxy mode",
			directMode:    false,
			host:          "http://localhost:9117",
			endpoint:      "all/results/torznab/api",
			params:        map[string]string{"t": "search", "q": "ubuntu"},
			expectedPath:  "/api/v2.0/indexers/all/results/torznab/api",
			expectedQuery: map[string]string{"t": "search", "q": "ubuntu"},
		},
		{
			name:          "direct mode with endpoint",
			directMode:    true,
			host:          "https://tracker.example.com/api/torznab",
			endpoint:      "search",
			params:        map[string]string{"t": "search"},
			expectedPath:  "/api/torznab/search",
			expectedQuery: map[string]string{"t": "search"},
		},
		{
			name:          "direct mode without endpoint",
			directMode:    true,
			host:          "https://tracker.example.com/api/torznab",
			endpoint:      "",
			params:        map[string]string{"t": "caps"},
			expectedPath:  "/api/torznab",
			expectedQuery: map[string]string{"t": "caps"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient(Config{
				Host:       tt.host,
				DirectMode: tt.directMode,
			})

			url := client.buildUrl(tt.endpoint, tt.params)
			assert.Contains(t, url, tt.expectedPath)
			for k, v := range tt.expectedQuery {
				assert.Contains(t, url, k+"="+v)
			}
		})
	}
}

func TestDrainAndClose(t *testing.T) {
	t.Run("nil body", func(t *testing.T) {
		// Should not panic
		drainAndClose(nil)
	})

	t.Run("valid body", func(t *testing.T) {
		body := io.NopCloser(strings.NewReader("test content"))
		drainAndClose(body)
		// Should not panic and should close the body
	})
}

func TestBasicAuth(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		assert.True(t, ok)
		assert.Equal(t, "testuser", user)
		assert.Equal(t, "testpass", pass)

		w.Header().Set("Content-Type", "application/xml")
		response := `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
	<channel>
		<item>
			<title>Test</title>
		</item>
	</channel>
</rss>`
		w.Write([]byte(response))
	}))
	defer server.Close()

	client := NewClient(Config{
		Host:      server.URL,
		APIKey:    "test-key",
		BasicUser: "testuser",
		BasicPass: "testpass",
	})

	_, err := client.SearchDirect("test", nil)
	require.NoError(t, err)
}

func TestRetryOn5xxError(t *testing.T) {
	attemptCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attemptCount++
		if attemptCount < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.Header().Set("Content-Type", "application/xml")
		response := `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
	<channel>
		<item>
			<title>Success</title>
		</item>
	</channel>
</rss>`
		w.Write([]byte(response))
	}))
	defer server.Close()

	client := NewClient(Config{
		Host:   server.URL,
		APIKey: "test-key",
	})

	results, err := client.SearchDirect("test", nil)
	require.NoError(t, err)
	assert.Len(t, results.Channel.Item, 1)
	assert.GreaterOrEqual(t, attemptCount, 3)
}

func TestInvalidXMLResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		w.Write([]byte("invalid xml content"))
	}))
	defer server.Close()

	client := NewClient(Config{
		Host:   server.URL,
		APIKey: "test-key",
	})

	_, err := client.GetIndexers()
	assert.Error(t, err)
}

func TestContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.Write([]byte("response"))
	}))
	defer server.Close()

	client := NewClient(Config{
		Host:   server.URL,
		APIKey: "test-key",
	})

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	_, err := client.GetIndexersCtx(ctx)
	assert.Error(t, err)
}
