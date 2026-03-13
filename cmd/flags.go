package cmd

import (
	"github.com/spf13/cobra"

	"github.com/mdjarv/db/internal/config"
	"github.com/mdjarv/db/internal/conn"
	"github.com/mdjarv/db/internal/db"
	_ "github.com/mdjarv/db/internal/db/postgres" // register driver
)

func init() {
	f := rootCmd.PersistentFlags()
	f.String("dsn", "", "full connection URL")
	f.StringP("connection", "c", "", "named connection from config")
	f.StringP("host", "H", "", "database host")
	f.IntP("port", "p", 5432, "database port")
	f.StringP("user", "U", "", "database user")
	f.StringP("dbname", "d", "", "database name")
	f.String("sslmode", "", "SSL mode")
	f.StringP("password", "W", "", "database password")
	f.String("theme", "", "color theme (e.g. default-dark, nord, dracula)")
}

func resolveConnection(cmd *cobra.Command) (conn.ConnectionConfig, error) {
	f := cmd.Flags()

	dsn, _ := f.GetString("dsn")
	connName, _ := f.GetString("connection")
	host, _ := f.GetString("host")
	port, _ := f.GetInt("port")
	user, _ := f.GetString("user")
	password, _ := f.GetString("password")
	dbname, _ := f.GetString("dbname")
	sslmode, _ := f.GetString("sslmode")

	if !f.Changed("port") {
		port = 0
	}

	stores := connectionStores()
	creds := conn.NewCredentialStore(conn.OSKeyring{})

	return conn.Resolve(conn.ResolveOptions{
		DSN:      dsn,
		ConnName: connName,
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
		DBName:   dbname,
		SSLMode:  sslmode,
	}, stores, creds)
}

// connectionStores returns [project, global] stores. Project store is nil
// when not inside a git repository.
func connectionStores() []*conn.Store {
	global := conn.NewStore(config.ConnectionsFile())
	if projFile := config.ProjectConnectionsFile(); projFile != "" {
		return []*conn.Store{conn.NewStore(projFile), global}
	}
	return []*conn.Store{global}
}

func connectFromFlags(cmd *cobra.Command) (db.Conn, error) {
	cfg, err := resolveConnection(cmd)
	if err != nil {
		return nil, err
	}
	return db.Open(cmd.Context(), "postgres", cfg.DSN())
}
