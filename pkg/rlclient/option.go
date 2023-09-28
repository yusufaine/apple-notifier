package rlclient

import (
	"net/http"
	"time"

	"golang.org/x/time/rate"
)

type RlOpts func(c *Client)

func WithHttpClient(cl *http.Client) RlOpts {
	return func(c *Client) {
		c.cl = cl
	}
}

func WithRateLimit(interval time.Duration, tokens int) RlOpts {
	return func(c *Client) {
		c.rl = rate.NewLimiter(rate.Every(interval), tokens)
	}
}
