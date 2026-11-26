package migrations

import (
	"context"
	"embed"
	"fmt"
	"io/fs"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Embed forward migrations into the Go executable so the runtime image
// does not need a separate copy of the SQL files.
//
//go:embed *.up.sql
var files embed.FS

// Apply runs pending migrations in filename order, in one transaction.
// Migration files must contain transactional SQL without BEGIN or COMMIT.
func Apply(ctx context.Context, pool *pgxpool.Pool) ([]string, error) {
	names, err := fs.Glob(files, "*.up.sql")
	if err != nil {
		return nil, fmt.Errorf("find migrations: %w", err)
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin migration transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Serialize cooperating migration runners, even across API instances.
	if _, err := tx.Exec(ctx, "SELECT pg_advisory_xact_lock(74632001)"); err != nil {
		return nil, fmt.Errorf("acquire migration lock: %w", err)
	}

	_, err = tx.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version TEXT PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`)
	if err != nil {
		return nil, fmt.Errorf("create migration history: %w", err)
	}

	var applied []string
	for _, name := range names {
		var exists bool
		err := tx.QueryRow(ctx,
			"SELECT EXISTS (SELECT 1 FROM schema_migrations WHERE version = $1)",
			name,
		).Scan(&exists)
		if err != nil {
			return nil, fmt.Errorf("check migration %s: %w", name, err)
		}
		if exists {
			continue
		}

		sql, err := files.ReadFile(name)
		if err != nil {
			return nil, fmt.Errorf("read migration %s: %w", name, err)
		}

		if _, err := tx.Exec(ctx, string(sql)); err != nil {
			return nil, fmt.Errorf("apply migration %s: %w", name, err)
		}

		if _, err := tx.Exec(ctx,
			"INSERT INTO schema_migrations (version) VALUES ($1)", name,
		); err != nil {
			return nil, fmt.Errorf("record migration %s: %w", name, err)
		}
		applied = append(applied, name)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit migrations: %w", err)
	}
	return applied, nil
}
