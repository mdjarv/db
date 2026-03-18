// Package dump wraps pg_dump for database dump with progress tracking.
package dump

import (
	"fmt"
	"time"
)

// Format represents a pg_dump output format.
type Format int

// Dump output formats.
const (
	Plain     Format = iota // SQL text
	Custom                  // compressed, supports pg_restore
	Directory               // directory with per-table files
	Tar                     // tar archive
)

func (f Format) String() string {
	switch f {
	case Plain:
		return "plain"
	case Custom:
		return "custom"
	case Directory:
		return "directory"
	case Tar:
		return "tar"
	default:
		return fmt.Sprintf("unknown(%d)", int(f))
	}
}

// flag returns the pg_dump --format flag value.
func (f Format) flag() string {
	switch f {
	case Plain:
		return "p"
	case Custom:
		return "c"
	case Directory:
		return "d"
	case Tar:
		return "t"
	default:
		return "c"
	}
}

// ext returns the file extension for this format.
func (f Format) ext() string {
	switch f {
	case Plain:
		return ".sql"
	case Custom:
		return ".dump"
	case Tar:
		return ".tar"
	case Directory:
		return "" // directory name, no extension
	default:
		return ".dump"
	}
}

// Config holds parameters for a pg_dump invocation.
type Config struct {
	Host       string
	Port       string
	User       string
	Password   string
	DBName     string
	SSLMode    string
	Format     Format
	SchemaOnly bool
	Tables     []string
	OutputPath string
}

// BuildArgs constructs pg_dump CLI arguments from a Config.
func BuildArgs(cfg Config) []string {
	var args []string

	if cfg.Host != "" {
		args = append(args, "--host", cfg.Host)
	}
	if cfg.Port != "" {
		args = append(args, "--port", cfg.Port)
	}
	if cfg.User != "" {
		args = append(args, "--username", cfg.User)
	}
	if cfg.DBName != "" {
		args = append(args, "--dbname", cfg.DBName)
	}

	args = append(args, "--format="+cfg.Format.flag())

	if cfg.SchemaOnly {
		args = append(args, "--schema-only")
	}

	for _, t := range cfg.Tables {
		args = append(args, "-t", t)
	}

	if cfg.OutputPath != "" {
		args = append(args, "-f", cfg.OutputPath)
	}

	args = append(args, "--verbose", "--no-password")

	return args
}

// DefaultOutputPath generates a default dump filename: <dbname>_<YYYYMMDD>.<ext>.
// For Directory format, no extension is appended.
func DefaultOutputPath(dbname string, format Format) string {
	date := time.Now().Format("20060102")
	return dbname + "_" + date + format.ext()
}
