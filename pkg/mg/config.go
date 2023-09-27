package mg

import (
	"context"
	"os"
)

type Config struct {
	Context  context.Context
	MongoDb  string
	MongoUri string
}

func (c *Config) mustValidate() {
	if c.MongoDb == "" {
		panic("'MONGO_DB' must be specified")
	}
	if c.MongoUri == "" {
		panic("'MONGO_URI' must be specified")
	}
}

func NewConfig(ctx context.Context) *Config {
	c := &Config{
		Context:  ctx,
		MongoDb:  os.Getenv("MONGO_DB"),
		MongoUri: os.Getenv("MONGO_URI"),
	}
	c.mustValidate()

	return c
}
