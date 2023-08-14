package jackett

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/autobrr/go-qbittorrent/errors"
	"github.com/avast/retry-go"
)

func (c *Client) getRawCtx(ctx context.Context, reqUrl string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqUrl, nil)
	if err != nil {
		return nil, errors.Wrap(err, "could not build request")
	}

	if c.cfg.BasicUser != "" && c.cfg.BasicPass != "" {
		req.SetBasicAuth(c.cfg.BasicUser, c.cfg.BasicPass)
	}

	// try request and if fail run 10 retries
	resp, err := c.retryDo(ctx, req)
	if err != nil {
		return nil, errors.Wrap(err, "error making get request: %v", reqUrl)
	}

	return resp, nil
}

func (c *Client) getCtx(ctx context.Context, endpoint string, opts map[string]string) (*http.Response, error) {
	return c.getRawCtx(ctx, c.buildUrl(endpoint, opts))
}

func (c *Client) postCtx(ctx context.Context, endpoint string, opts map[string]string) (*http.Response, error) {
	// add optional parameters that the user wants
	form := url.Values{}
	for k, v := range opts {
		form.Add(k, v)
	}

	reqUrl := c.buildUrl(endpoint, nil)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqUrl, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, errors.Wrap(err, "could not build request")
	}

	if c.cfg.BasicUser != "" && c.cfg.BasicPass != "" {
		req.SetBasicAuth(c.cfg.BasicUser, c.cfg.BasicPass)
	}

	// add the content-type so qbittorrent knows what to expect
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	// try request and if fail run 10 retries
	resp, err := c.retryDo(ctx, req)
	if err != nil {
		return nil, errors.Wrap(err, "error making post request: %v", reqUrl)
	}

	return resp, nil
}

func (c *Client) postBasicCtx(ctx context.Context, endpoint string, opts map[string]string) (*http.Response, error) {
	// add optional parameters that the user wants
	form := url.Values{}
	for k, v := range opts {
		form.Add(k, v)
	}

	var resp *http.Response

	reqUrl := c.buildUrl(endpoint, nil)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqUrl, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, errors.Wrap(err, "could not build request")
	}

	if c.cfg.BasicUser != "" && c.cfg.BasicPass != "" {
		req.SetBasicAuth(c.cfg.BasicUser, c.cfg.BasicPass)
	}

	// add the content-type so qbittorrent knows what to expect
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err = c.http.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "error making post request: %v", reqUrl)
	}

	return resp, nil
}

func (c *Client) setCookies(cookies []*http.Cookie) {
	cookieURL, _ := url.Parse(c.buildUrl("/", nil))

	c.http.Jar.SetCookies(cookieURL, cookies)
}

func (c *Client) buildUrl(endpoint string, params map[string]string) string {
	apiBase := "/api/v2.0/indexers/"

	// add query params
	queryParams := url.Values{}
	for key, value := range params {
		queryParams.Add(key, value)
	}

	joinedUrl, _ := url.JoinPath(c.cfg.Host, apiBase, endpoint)
	parsedUrl, _ := url.Parse(joinedUrl)
	parsedUrl.RawQuery = queryParams.Encode()

	// make into new string and return
	return parsedUrl.String()
}

func copyBody(src io.ReadCloser) ([]byte, error) {
	b, err := io.ReadAll(src)
	if err != nil {
		// ErrReadingRequestBody
		return nil, err
	}
	src.Close()
	return b, nil
}

func resetBody(request *http.Request, originalBody []byte) {
	request.Body = io.NopCloser(bytes.NewBuffer(originalBody))
	request.GetBody = func() (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewBuffer(originalBody)), nil
	}
}

func (c *Client) retryDo(ctx context.Context, req *http.Request) (*http.Response, error) {
	var (
		originalBody []byte
		err          error
	)

	if req != nil && req.Body != nil {
		originalBody, err = copyBody(req.Body)
		resetBody(req, originalBody)
	}

	if err != nil {
		return nil, err
	}

	var resp *http.Response

	// try request and if fail run 10 retries
	err = retry.Do(func() error {
		resp, err = c.http.Do(req)

		if err == nil {
			if resp.StatusCode < 500 {
				return err
			} else if resp.StatusCode >= 500 {
				return retry.Unrecoverable(errors.New("unrecoverable status: %v", resp.StatusCode))
			}
		}

		retry.Delay(time.Second * 3)

		return err
	},
		retry.OnRetry(func(n uint, err error) { c.log.Printf("%q: attempt %d - %v\n", err, n, req.URL.String()) }),
		//retry.Delay(time.Second*3),
		retry.Attempts(5),
		retry.MaxJitter(time.Second*1),
	)

	if err != nil {
		return nil, errors.Wrap(err, "error making request")
	}

	return resp, nil
}
