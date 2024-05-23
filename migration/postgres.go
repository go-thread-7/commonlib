package migration

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/go-thread-7/commonlib/migration/config"
	"github.com/go-thread-7/commonlib/migration/contracts"

	"emperror.dev/errors"
	"github.com/golang-migrate/migrate"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

type goMigratePostgresMigrator struct {
	config     *config.MigrationOptions
	db         *sql.DB
	datasource string
	migration  *migrate.Migrate
}

func NewGoMigratorPostgres(config *config.MigrationOptions, db *sql.DB) (contracts.PostgresMigrationRunner, error) {
	datasource := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		config.User,
		config.Password,
		config.Host,
		config.Port,
		config.DBName,
	)

	migration, err := migrate.New(fmt.Sprintf("file://%s", config.MigrationsDir), datasource)
	if err != nil {
		return nil, errors.WrapIf(err, "failed to initialize migrator")
	}

	return &goMigratePostgresMigrator{
		config:     config,
		db:         db,
		datasource: datasource,
		migration:  migration,
	}, nil
}

func (m *goMigratePostgresMigrator) Up(_ context.Context, version uint) error {
	if m.config.SkipMigration {
		log.Println("database migration skipped")
		return nil
	}

	err := m.executeCommand(config.Up, version)
	if errors.Is(err, migrate.ErrNoChange) {
		return nil
	}
	if err != nil {
		return errors.WrapIf(err, "failed to migrate database")
	}

	log.Println("migration finished")

	return nil
}

func (m *goMigratePostgresMigrator) Down(_ context.Context, version uint) error {
	if m.config.SkipMigration {
		log.Println("database migration skipped")
		return nil
	}

	err := m.executeCommand(config.Down, version)
	if errors.Is(err, migrate.ErrNoChange) {
		return nil
	}
	if err != nil {
		return errors.WrapIf(err, "failed to migrate database")
	}

	log.Println("migration finished")

	return nil
}

func (m *goMigratePostgresMigrator) executeCommand(command config.CommandType, version uint) error {
	var err error
	switch command {
	case config.Up:
		if version == 0 {
			err = m.migration.Up()
		} else {
			err = m.migration.Migrate(version)
		}
	case config.Down:
		if version == 0 {
			err = m.migration.Down()
		} else {
			err = m.migration.Migrate(version)
		}
	default:
		err = errors.New("invalid migration direction")
	}

	if err == migrate.ErrNoChange {
		return nil
	}
	if err != nil {
		return errors.WrapIf(err, "failed to migrate database")
	}

	return nil
}
