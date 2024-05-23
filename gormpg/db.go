package gormpg

import (
	"fmt"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/go-thread-7/commonlib/gormpg/config"
	"github.com/pkg/errors"
	gorm_postgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func New(config *config.GORMPostgresConfig) (*gorm.DB, error) {

	bo := backoff.NewExponentialBackOff()
	bo.MaxElapsedTime = 10 * time.Second
	maxRetries := 5

	dsn := fmt.Sprintf("host=%s port=%d user=%s dbname=%s password=%s",
		config.Host,
		config.Port,
		config.User,
		config.DBName,
		config.Password,
	)

	var db *gorm.DB
	var err error
	err = backoff.Retry(func() error {
		db, err = gorm.Open(gorm_postgres.Open(dsn), &gorm.Config{})
		if err != nil {
			return errors.Errorf("failed to connect postgres: %v and connection information: %s", err, dsn)
		}
		return nil
	}, backoff.WithMaxRetries(bo, uint64(maxRetries-1)))

	return db, err
}
