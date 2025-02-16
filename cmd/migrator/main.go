package main

import (
	"errors"
	"flag"
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	dbURL, migrationsPath, action, version := parseFlags()

	m, err := newMigrateInstance(*dbURL, *migrationsPath)
	if err != nil {
		log.Fatalf("Could not create migrate instance: %v", err)
	}

	if err := executeMigration(m, *action, *version); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	fmt.Println("Migration successful!")
}

func parseFlags() (*string, *string, *string, *int) {
	dbURL := flag.String("dburl", "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable", "Database connection URL")
	migrationsPath := flag.String("migrationspath", "migrations", "Path to migrations folder")
	action := flag.String("action", "up", "Specify migration action: up, down, force, version")
	version := flag.Int("version", -1, "Specify migration version for force (only for -action=force)")

	flag.Parse()
	return dbURL, migrationsPath, action, version
}

func newMigrateInstance(dbURL, migrationsPath string) (*migrate.Migrate, error) {
	return migrate.New(fmt.Sprintf("file://%s", migrationsPath), dbURL)
}

func executeMigration(m *migrate.Migrate, action string, version int) error {
	switch action {
	case "up":
		return handleMigrationError(m.Up())
	case "down":
		return handleMigrationError(m.Down())
	case "force":
		if version < 0 {
			return errors.New("Error: You must provide a valid version number with -version for force action")
		}
		return handleMigrationError(m.Force(version))
	case "version":
		return printMigrationVersion(m)
	default:
		fmt.Println("Usage: migrate -action=[up|down|force|version] -dburl=<DB_URL> -migrationspath=<path> -version=<number>")
		return nil
	}
}

func handleMigrationError(err error) error {
	if err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			fmt.Println("No new migrations to apply.")
			return nil
		}
		return err
	}
	return nil
}

func printMigrationVersion(m *migrate.Migrate) error {
	version, dirty, err := m.Version()
	if err != nil && !errors.Is(err, migrate.ErrNilVersion) {
		return fmt.Errorf("Could not get migration version: %v", err)
	}
	fmt.Printf("Current migration version: %d (dirty: %t)\n", version, dirty)
	return nil
}
