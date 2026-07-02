package repository

import (
	"database/sql"
	"fmt"
	"io/fs"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

// i didn't directly use sql.DB everywhere bcz it is possiblr that at a later stage i would want to may be add a mutex which may break the app so just thought of keeping a low cost abstraction for future usability
type DB struct {
	*sql.DB
}

// inits the DB and makes the migs actually run
func NewDB(dburl string, migfs fs.FS) (*DB, error) {
	db, err := sql.Open("pgx", dburl)
	if err != nil {
		return nil, fmt.Errorf("db not opening: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("db didn't receive/respond to the ping: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	if err := run_migs(db, migfs); err != nil {
		return nil, fmt.Errorf("migs couldn't be completed %w", err)
	}
	return &DB{db}, nil
}

// utility func to define how to run the migs carefully
func run_migs(db *sql.DB, migfs fs.FS) error {
	drvr, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("cdn't create mig server: %w", err)
	}
	src, err := iofs.New(migfs, ".")
	if err != nil {
		return fmt.Errorf("iofs src couldnt be created: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", src, "postgres", drvr)

	if err := m.Up(); err != nil {
		return fmt.Errorf("migration oculd not be run up: %w", err)
	}
	return nil
}
