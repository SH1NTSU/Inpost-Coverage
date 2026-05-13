package httpx

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Client struct {
	HTTP    *http.Client
	BaseURL string
	UA      string
	Limiter Limiter
}

func New(baseURL string, timeout time.Duration, limiter Limiter) *Client {
	return &Client{
		HTTP:    &http.Client{Timeout: timeout},
		BaseURL: baseURL,
		UA:      "paczkomat-predictor/0.1",
		Limiter: limiter,
	}
}

func (c *Client) Do(ctx context.Context, method, path string, body io.Reader, out any) error {
	if c.Limiter != nil {
		if err := c.Limiter.Wait(ctx); err != nil {
			return err
		}
	}
	req, err := http.NewRequestWithContext(ctx, method, c.BaseURL+path, body)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", c.UA)
	req.Header.Set("Accept", "application/json")
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		buf, _ := io.ReadAll(io.LimitReader(resp.Body, 4<<10))
		return fmt.Errorf("http %d: %s", resp.StatusCode, string(buf))
	}
	if out == nil {
		_, _ = io.Copy(io.Discard, resp.Body)
		return nil
	}
	return json.NewDecoder(resp.Body).Decode(out)
}
