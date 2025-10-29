package database

import (
    "io/fs"
    "path/filepath"
    "sort"
    "strconv"
    "strings"
    "testing"
)

var migrationDirectories = map[string]string{
    "sqlite":   "migrations/sqlite",
    "postgres": "migrations/postgres",
}

func TestMigrationFilesHavePairs(t *testing.T) {
    for driver, dir := range migrationDirectories {
        driver := driver
        dir := dir
        t.Run(driver, func(t *testing.T) {
            entries, err := fs.ReadDir(migrationFiles, dir)
            if err != nil {
                t.Fatalf("failed to read directory %s: %v", dir, err)
            }

            if len(entries) == 0 {
                t.Fatalf("no migration files found in %s", dir)
            }

            type directions struct {
                up   bool
                down bool
            }

            versions := map[uint64]*directions{}

            for _, entry := range entries {
                if entry.IsDir() {
                    continue
                }

                name := entry.Name()
                if !strings.HasSuffix(name, ".sql") {
                    t.Fatalf("unexpected non-SQL file %s", name)
                }

                parts := strings.SplitN(name, "_", 2)
                if len(parts) != 2 {
                    t.Fatalf("invalid migration file name: %s", name)
                }

                version, err := strconv.ParseUint(parts[0], 10, 64)
                if err != nil {
                    t.Fatalf("invalid migration version in %s: %v", name, err)
                }

                suffix := parts[1]
                switch {
                case strings.HasSuffix(suffix, ".up.sql"):
                    v := versions[version]
                    if v == nil {
                        v = &directions{}
                        versions[version] = v
                    }
                    if v.up {
                        t.Fatalf("duplicate up migration for version %d", version)
                    }
                    v.up = true
                case strings.HasSuffix(suffix, ".down.sql"):
                    v := versions[version]
                    if v == nil {
                        v = &directions{}
                        versions[version] = v
                    }
                    if v.down {
                        t.Fatalf("duplicate down migration for version %d", version)
                    }
                    v.down = true
                default:
                    t.Fatalf("migration file must end with .up.sql or .down.sql: %s", name)
                }
            }

            if len(versions) == 0 {
                t.Fatalf("no valid migrations detected in %s", dir)
            }

            var keys []uint64
            for version, dirs := range versions {
                if !dirs.up || !dirs.down {
                    t.Fatalf("version %d in %s must have both up and down files", version, dir)
                }
                keys = append(keys, version)
            }

            sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })

            if keys[0] != 1 {
                t.Fatalf("first migration version must be 000001, found %06d", keys[0])
            }

            for i := 1; i < len(keys); i++ {
                if keys[i] != keys[i-1]+1 {
                    t.Fatalf("migration versions must be contiguous; found gap between %d and %d", keys[i-1], keys[i])
                }
            }

            // Quick sanity check: ensure files are accessible from embed path as well
            // Ensure files are embedded
            for _, entry := range entries {
                if entry.IsDir() {
                    continue
                }
                path := filepath.Join("migrations", driver, entry.Name())
                f, err := migrationFiles.Open(path)
                if err != nil {
                    t.Fatalf("embedded migration missing: %s (%v)", path, err)
                }
                f.Close()
            }
        })
    }
}
