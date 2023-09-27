package tg

import (
	"context"
	"os"
)

type Config struct {
	ChatId  string
	Context context.Context
	Token   string
}

func (c *Config) mustValidate() {
	if c.ChatId == "" {
		panic("'TG_CHAT_ID' must be specified")
	}
	if c.Token == "" {
		panic("'TG_BOT_TOKEN' must be specified")
	}
}

func NewConfig(ctx context.Context) *Config {
	c := &Config{
		ChatId:  os.Getenv("TG_CHAT_ID"),
		Context: ctx,
		Token:   os.Getenv("TG_BOT_TOKEN"),
	}
	c.mustValidate()

	return c
}
