//go:build integration

package dump

import (
	"context"
	"testing"
)

func TestRunnerIntegration(t *testing.T) {
	binary, err := FindPgDump("")
	if err != nil {
		t.Skip("pg_dump not available:", err)
	}

	runner := NewRunner(binary)
	cfg := Config{
		Host:       "localhost",
		Port:       "5432",
		User:       "postgres",
		DBName:     "postgres",
		Format:     Plain,
		SchemaOnly: true,
		OutputPath: t.TempDir() + "/test.sql",
	}

	ch, err := runner.Run(context.Background(), cfg, 0)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	for ev := range ch {
		if ev.Err != nil {
			t.Logf("pg_dump error: %v", ev.Err)
		}
	}
}
