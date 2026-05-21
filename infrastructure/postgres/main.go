package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/url"

	"github.com/go-pg/pg/v10"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	postgresDriver "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

type postgres struct {
	db *pg.DB
}

var instance *postgres

func migration() error {
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=disable",
		cfg.Username,
		url.QueryEscape(cfg.Password),
		cfg.Host,
		cfg.Port,
		cfg.Database,
	)
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return err
	}
	defer db.Close()

	driver, err := postgresDriver.WithInstance(db, &postgresDriver.Config{
		MigrationsTable: "taskforge_schema_migrations",
	})
	if err != nil {
		return err
	}

	m, err := migrate.NewWithDatabaseInstance(cfg.MigrationPath, "postgres", driver)
	if err != nil {
		return err
	}

	err = m.Up()
	if err == migrate.ErrNoChange {
		log.Println("no change since the last migration")
		return nil
	}
	if err != nil && err != migrate.ErrNoChange {
		return err
	}
	log.Println("migration completed")
	return nil
}

func NewPostgres() (*postgres, error) {
	if err := cfg.Validation(); err != nil {
		return nil, err
	}
	if err := migration(); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInitiatePostgres, err)
	}

	p := &postgres{}
	p.db = pg.Connect(&pg.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		User:     cfg.Username,
		Password: cfg.Password,
		Database: cfg.Database,
	})
	if _, err := p.db.Exec("SELECT 1"); err != nil {
		_ = p.db.Close()
		return nil, fmt.Errorf("%w: %v", ErrInitiatePostgres, err)
	}
	instance = p
	return p, nil
}

func (p *postgres) Close() error {
	if p == nil || p.db == nil {
		return nil
	}
	return p.db.Close()
}

func Ping(ctx context.Context) error {
	if instance == nil || instance.db == nil {
		return ErrPostgresNotReady
	}
	if _, err := instance.db.WithContext(ctx).Exec("SELECT 1"); err != nil {
		return ErrPostgresNotReady
	}
	return nil
}
