package configs

import (
	"time"

	"github.com/CXTACLYSM/postgres-ha-practice/configs/app"
	"github.com/CXTACLYSM/postgres-ha-practice/configs/database/persistence/postgres"
	"github.com/spf13/viper"
)

type Config struct {
	App             *app.Config
	PostgresCluster *postgres.ClusterConfig
}

func Create() *Config {
	viper.AutomaticEnv()

	return &Config{
		App: &app.Config{
			Host: viper.GetString("APP_HOST"),
			Http: app.Http{
				Port:              viper.GetInt("APP_HTTP_PORT"),
				ReadHeaderTimeout: 5 * time.Second,
				ReadTimeout:       10 * time.Second,
				WriteTimeout:      35 * time.Second,
				IdleTimeout:       60 * time.Second,
				MaxHeaderBytes:    1 << 20,
			},
		},
		PostgresCluster: &postgres.ClusterConfig{
			Read: &postgres.Config{
				Host:     viper.GetString("POSTGRES_READ_HOST"),
				Port:     viper.GetInt("POSTGRES_READ_PORT"),
				Username: viper.GetString("POSTGRES_READ_USERNAME"),
				Password: viper.GetString("POSTGRES_READ_PASSWORD"),
				Database: viper.GetString("POSTGRES_READ_DATABASE"),
			},
			Write: &postgres.Config{
				Host:     viper.GetString("POSTGRES_WRITE_HOST"),
				Port:     viper.GetInt("POSTGRES_WRITE_PORT"),
				Username: viper.GetString("POSTGRES_WRITE_USERNAME"),
				Password: viper.GetString("POSTGRES_WRITE_PASSWORD"),
				Database: viper.GetString("POSTGRES_WRITE_DATABASE"),
			},
		},
	}
}
