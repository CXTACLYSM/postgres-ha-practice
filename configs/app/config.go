package app

import (
	"fmt"
	"time"
)

type Config struct {
	Host string
	Http Http
}

type Http struct {
	Port              int
	ReadHeaderTimeout time.Duration
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
	MaxHeaderBytes    int
}

func (c *Config) HttpSocketStr() string {
	return fmt.Sprintf("%s:%d", "0.0.0.0", c.Http.Port)
}
