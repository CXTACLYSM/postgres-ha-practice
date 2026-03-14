package postgres

import (
	"errors"
	"fmt"

	"github.com/CXTACLYSM/postgres-ha-practice/pkg/postgres"
)

type ClusterConfig struct {
	Read  *Config
	Write *Config
}

type Config struct {
	Host     string
	Port     int
	Username string
	Password string
	Database string
}

func (c *ClusterConfig) DSN(operation int) (string, error) {
	switch operation {
	case postgres.ReadOperation:
		return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			c.Read.Host, c.Read.Port, c.Read.Username, c.Read.Password, c.Read.Database,
		), nil
	case postgres.WriteOperation:
		return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			c.Write.Host, c.Write.Port, c.Write.Username, c.Write.Password, c.Write.Database,
		), nil
	default:
		return "", errors.New("invalid pool type")
	}
}
