package rlclient

import (
	"net/http"

	"golang.org/x/time/rate"
)

type RlOpts func(c *Client)

func WithHttpClient(cl *http.Client) RlOpts {
	return func(c *Client) {
		c.cl = cl
	}
}

func WithRateLimiter(rl *rate.Limiter) RlOpts {
	return func(c *Client) {
		c.rl = rl
	}
}
