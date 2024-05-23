package gormpgsql

import (
	"fmt"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/pkg/errors"
	gorm_postgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type GormPostgresConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	DBName   string `mapstructure:"dbName"`
	SSLMode  bool   `mapstructure:"sslMode"`
	Password string `mapstructure:"password"`
}

func New(config *GormPostgresConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s dbname=%s password=%s",
		config.Host,
		config.Port,
		config.User,
		config.DBName,
		config.Password,
	)

	bo := backoff.NewExponentialBackOff()
	bo.MaxElapsedTime = 10 * time.Second
	maxRetries := 5

	var gormDb *gorm.DB
	var err error
	err = backoff.Retry(func() error {
		gormDb, err = gorm.Open(gorm_postgres.Open(dsn), &gorm.Config{})
		if err != nil {
			return errors.Errorf("failed to connect postgres: %v and connection information: %s", err, dsn)
		}
		return nil
	}, backoff.WithMaxRetries(bo, uint64(maxRetries-1)))

	return gormDb, err
}
