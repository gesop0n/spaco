package app

import (
	"testing"

	"github.com/jackc/pgx/v5"
)

func TestDatabasePoolConfigUsesExecMode(t *testing.T) {
	t.Parallel()

	config, err := databasePoolConfig(
		"postgresql://postgres:password@localhost:5432/spaco?sslmode=disable",
	)
	if err != nil {
		t.Fatalf("databasePoolConfig() error = %v", err)
	}
	if config.ConnConfig.DefaultQueryExecMode != pgx.QueryExecModeExec {
		t.Fatalf(
			"DefaultQueryExecMode = %v, want %v",
			config.ConnConfig.DefaultQueryExecMode,
			pgx.QueryExecModeExec,
		)
	}
}
