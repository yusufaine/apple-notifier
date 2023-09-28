package mongodb

import (
	"context"
	"os"
)

type Config struct {
	Context   context.Context
	MongoColl string
	MongoDb   string
	MongoUri  string
}

func (c *Config) mustValidate() {
	if c.MongoColl == "" {
		panic("'MONGO_COLL' must be specified")
	}
	if c.MongoDb == "" {
		panic("'MONGO_DB' must be specified")
	}
	if c.MongoUri == "" {
		panic("'MONGO_URI' must be specified")
	}
}

func NewConfig(ctx context.Context) *Config {
	c := &Config{
		Context:   ctx,
		MongoColl: os.Getenv("MONGO_COLL"),
		MongoDb:   os.Getenv("MONGO_DB"),
		MongoUri:  os.Getenv("MONGO_URI"),
	}
	c.mustValidate()

	return c
}
