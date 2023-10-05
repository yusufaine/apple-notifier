package rlclient

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"golang.org/x/time/rate"
)

type Client struct {
	context.Context
	cl *http.Client
	rl *rate.Limiter
}

// Creates a new rate-limited client, defaulting to 1 request per second,
// and is configurable using RlOpts.
func New(ctx context.Context, opts ...RlOpts) *Client {
	client := &Client{
		Context: ctx,
		cl:      http.DefaultClient,
		rl:      rate.NewLimiter(1, 1),
	}

	for _, opt := range opts {
		opt(client)
	}

	return client
}

func (c *Client) Do(req *http.Request) ([]byte, error) {
	if err := c.rl.Wait(c.Context); err != nil {
		return nil, err
	}

	resp, err := c.cl.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp == nil {
		return nil, fmt.Errorf("response body is nil")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf(string(body))
	}

	return body, nil
}
