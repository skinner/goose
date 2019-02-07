package goose

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"time"
)

// Status prints the status of all migrations.
func Status(db *sql.DB, dir string) error {
	// collect all migrations
	migrations, err := CollectMigrations(dir, minVersion, maxVersion)
	if err != nil {
		return err
	}

	// Gets current version and ensures that the version table exists if we're
	// running on a pristine DB.
	currentVersion, err := EnsureDBVersion(db)
	if err != nil {
		return err
	}

	if len(migrations) == 0 {
		fmt.Printf("goose: no migrations. current version: %d\n", currentVersion)
		return nil
	} else {
		fmt.Printf("goose: finding migrations. current version: %d\n", currentVersion)
	}

	log.Println("    Applied At                  Migration")
	log.Println("    =======================================")
	for _, migration := range migrations {
		printMigrationStatus(db, migration.Version, filepath.Base(migration.Source), currentVersion)
	}

	return nil
}

func printMigrationStatus(db *sql.DB, version int64, script string, currentVersion int64) {
	var row MigrationRecord
	q := fmt.Sprintf("SELECT tstamp, is_applied FROM %s WHERE version_id=%d ORDER BY tstamp DESC LIMIT 1", TableName(), version)
	e := db.QueryRow(q).Scan(&row.TStamp, &row.IsApplied)

	if e != nil && e != sql.ErrNoRows {
		log.Fatal(e)
	}

	var appliedAt string

	if row.IsApplied {
		appliedAt = row.TStamp.Format(time.ANSIC)
	} else {
		if version < currentVersion {
			appliedAt = "Pending (MISSED)"
		} else {
			appliedAt = "Pending"
		}
	}

	log.Printf("    %-24s -- %v\n", appliedAt, script)
}

// StatusUnapplied prints all unapplied migrations
func StatusUnapplied(db *sql.DB, dir string) error {
	migrations, err := CollectUnappliedMigrations(db, dir)
	if err != nil {
		return err
	}

	// Gets current version and ensures that the version table exists if we're
	// running on a pristine DB.
	currentVersion, err := EnsureDBVersion(db)
	if err != nil {
		return err
	}

	if len(migrations) == 0 {
		fmt.Printf("goose: no unapplied migrations. current version: %d\n", currentVersion)
		return nil
	} else {
		fmt.Printf("goose: finding unapplied migrations. current version: %d\n", currentVersion)
	}

	fmt.Println("    Unapplied migrations")
	fmt.Println("    ===========")
	for _, migration := range migrations {
		if migration.Version < currentVersion {
			fmt.Printf("    %v (MISSED)\n", filepath.Base(migration.Source))
		} else {
			fmt.Printf("    %v\n", filepath.Base(migration.Source))
		}
	}

	return nil
}
