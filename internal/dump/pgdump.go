package dump

import (
	"fmt"
	"os/exec"
	"strings"
)

// FindPgDump locates the pg_dump binary. If configPath is non-empty, it is
// used directly. Otherwise exec.LookPath searches PATH.
func FindPgDump(configPath string) (string, error) {
	if configPath != "" {
		return configPath, nil
	}
	path, err := exec.LookPath("pg_dump")
	if err != nil {
		return "", fmt.Errorf("pg_dump not found in PATH; install postgresql-client or set pg_dump_path in config")
	}
	return path, nil
}

// PgDumpVersion runs the binary with --version and returns the version string.
func PgDumpVersion(binary string) (string, error) {
	out, err := exec.Command(binary, "--version").Output()
	if err != nil {
		return "", fmt.Errorf("pg_dump --version: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}
