package jackett

import (
	"context"
	"encoding/xml"
	"io"

	"github.com/autobrr/go-qbittorrent/errors"
)

func (c *Client) GetIndexers() (Indexers, error) {
	return c.GetIndexersCtx(context.Background())
}

func (c *Client) GetIndexersCtx(ctx context.Context) (Indexers, error) {
	opts := map[string]string{
		"t":          "indexers",
		"configured": "true",
	}

	if len(c.cfg.APIKey) != 0 {
		opts["apikey"] = c.cfg.APIKey
	}

	var ind Indexers
	resp, err := c.getCtx(ctx, "all/results/torznab/api", opts)
	if err != nil {
		return ind, errors.Wrap(err, "all endpoint error")
	}

	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return ind, err
	}

	err = xml.Unmarshal(bodyBytes, &ind)
	return ind, err
}

func (c *Client) GetTorrents(indexer string, opts map[string]string) (Rss, error) {
	return c.GetTorrentsCtx(context.Background(), indexer, opts)
}

func (c *Client) GetTorrentsCtx(ctx context.Context, indexer string, opts map[string]string) (Rss, error) {
	if len(c.cfg.APIKey) != 0 {
		opts["apikey"] = c.cfg.APIKey
	}

	var rss Rss
	resp, err := c.getCtx(ctx, indexer+"/results/torznab/api", opts)
	if err != nil {
		return rss, errors.Wrap(err, indexer+" endpoint error")
	}

	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return rss, err
	}

	err = xml.Unmarshal(bodyBytes, &rss)
	return rss, err
}
