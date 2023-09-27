package mg

import "os"

type Config struct {
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

func NewConfig() *Config {
	c := &Config{
		MongoDb:  os.Getenv("MONGO_DB"),
		MongoUri: os.Getenv("MONGO_URI"),
	}
	c.mustValidate()

	return c
}
