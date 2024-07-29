package jackett

import (
	"bufio"
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

	body := bufio.NewReader(resp.Body)
	if _, err := body.Peek(0); err != nil && err != bufio.ErrBufferFull {
		return ind, errors.Wrap(err, "unable to read body")
	}

	err = xml.NewDecoder(body).Decode(&ind)
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

	body := bufio.NewReader(resp.Body)
	if _, err := body.Peek(0); err != nil && err != bufio.ErrBufferFull {
		return rss, errors.Wrap(err, "unable to read body")
	}

	err = xml.NewDecoder(body).Decode(&rss)
	return rss, err
}

func (c *Client) GetEnclosure(enclosure string) ([]byte, error) {
	return c.GetEnclosureCtx(context.Background(), enclosure)
}

func (c *Client) GetEnclosureCtx(ctx context.Context, enclosure string) ([]byte, error) {
	resp, err := c.getRawCtx(ctx, enclosure)
	if err != nil {
		return nil, errors.Wrap(err, enclosure)
	}

	defer resp.Body.Close()

	return io.ReadAll(bufio.NewReader(resp.Body))
}
