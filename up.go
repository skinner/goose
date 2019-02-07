package goose

import (
	"database/sql"
	"fmt"
	"path/filepath"
)

// Up performs all types of migrations upwards, depending on the params.
func Up(db *sql.DB, dir string, includeMissing bool, onlyOne bool, endVersion *int64, dryRun bool) error {
	var migrations Migrations
	currentVersion, err := GetDBVersion(db)
	if err != nil {
		return err
	}
	unappliedMigrations, err := CollectUnappliedMigrations(db, dir)
	if err != nil {
		return err
	}
	if includeMissing {
		migrations = unappliedMigrations
	} else {
		migrations, err = CollectMigrations(dir, currentVersion, maxVersion)
		if err != nil {
			return err
		}
		if len(unappliedMigrations) != len(migrations) {
			return fmt.Errorf("missing migrations found! please run goose status -show-unapplied-only to find missing migrations, or run this command again with the -include-missing flag to apply them as well")
		}
	}
	if dryRun {
		log.Printf("goose: dry run. the following migrations would be applied. current version: %d\n", currentVersion)
	} else {
		log.Printf("goose: applying migrations. current version: %d\n", currentVersion)
	}
	statuses, err := dbMigrationsStatus(db)
	if err != nil {
		return err
	}
	for _, migration := range migrations {
		if endVersion != nil && migration.Version > *endVersion {
			break
		}
		if _, ok := statuses[migration.Version]; ok {
			log.Printf("goose version was out of sync. skipping already-applied migration %v\n", filepath.Base(migration.Source))
			continue
		}
		if dryRun {
			log.Println(filepath.Base(migration.Source))
		} else {
			if err := migration.Up(db); err != nil {
				return err
			}
		}
		if onlyOne {
			break
		}
	}
	currentVersion, err = GetDBVersion(db)
	if err != nil {
		return err
	}
	if dryRun {
		log.Println("goose: no more migrations would be run.")
	} else {
		log.Printf("goose: no more migrations to apply. current version: %d\n", currentVersion)
	}
	return nil
}
