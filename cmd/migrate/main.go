package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/kitakitabauer/gin-sample-app/config"
	"github.com/kitakitabauer/gin-sample-app/internal/database"
)

func main() {
	var (
		command string
		steps   int
		timeout time.Duration
	)

	flag.StringVar(&command, "cmd", "up", "migration command: up, down, steps, version")
	flag.IntVar(&steps, "steps", 0, "number of steps to migrate (used with cmd=steps or cmd=down)")
	flag.DurationVar(&timeout, "timeout", 5*time.Second, "database connection timeout")
	flag.Parse()

	config.Load()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	db, err := database.Open(ctx, database.Config{
		Driver: config.AppConfig.DatabaseDriver,
		DSN:    config.AppConfig.DatabaseDSN,
	})
	cancel()
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}
	defer db.Close()

	switch command {
	case "up":
		if err := database.MigrateUp(db, config.AppConfig.DatabaseDriver); err != nil {
			log.Fatalf("migrate up failed: %v", err)
		}
		fmt.Fprintln(os.Stdout, "migrations applied")
	case "down":
		if steps > 0 {
			steps = -steps
		}
		if steps != 0 {
			if err := database.MigrateSteps(db, config.AppConfig.DatabaseDriver, steps); err != nil {
				log.Fatalf("migrate steps failed (%T): %v", err, err)
			}
		} else {
			if err := database.MigrateDown(db, config.AppConfig.DatabaseDriver); err != nil {
				log.Fatalf("migrate down failed: %v", err)
			}
		}
		fmt.Fprintln(os.Stdout, "migrations rolled back")
	case "steps":
		if steps == 0 {
			log.Fatal("steps command requires --steps to be non-zero")
		}
		before, dirty, err := database.MigrationVersion(db, config.AppConfig.DatabaseDriver)
		if err != nil {
			log.Fatalf("failed to fetch current version: %v", err)
		}
		if dirty {
			log.Fatal("cannot run steps: database is in dirty state")
		}

		if err := database.MigrateSteps(db, config.AppConfig.DatabaseDriver, steps); err != nil {
			log.Fatalf("migrate steps failed (%T): %v", err, err)
		}

		after, dirty, err := database.MigrationVersion(db, config.AppConfig.DatabaseDriver)
		if err != nil {
			log.Fatalf("failed to fetch updated version: %v", err)
		}
		if dirty {
			log.Fatal("migration ended in dirty state")
		}

		if before == after {
			fmt.Fprintln(os.Stdout, "no migrations applied")
		} else {
			fmt.Fprintf(os.Stdout, "migrated from version %d to %d\n", before, after)
		}
	case "version":
		version, dirty, err := database.MigrationVersion(db, config.AppConfig.DatabaseDriver)
		if err != nil {
			log.Fatalf("fetching migration version failed: %v", err)
		}
		fmt.Fprintf(os.Stdout, "version=%d dirty=%t\n", version, dirty)
	default:
		log.Fatalf("unsupported command: %s", command)
	}
}
