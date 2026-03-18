package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/mdjarv/db/internal/config"
	"github.com/mdjarv/db/internal/conn"
)

var connectCmd = &cobra.Command{
	Use:   "connect",
	Short: "Manage saved connections",
}

var connectAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a saved connection",
	RunE:  runConnectAdd,
}

var connectListCmd = &cobra.Command{
	Use:   "list",
	Short: "List saved connections",
	RunE:  runConnectList,
}

var connectRemoveCmd = &cobra.Command{
	Use:   "remove <name>",
	Short: "Remove a saved connection",
	Args:  cobra.ExactArgs(1),
	RunE:  runConnectRemove,
}

var connectDefaultCmd = &cobra.Command{
	Use:   "default <name>",
	Short: "Set the default connection",
	Args:  cobra.ExactArgs(1),
	RunE:  runConnectDefault,
}

var connectRenameCmd = &cobra.Command{
	Use:   "rename <old> <new>",
	Short: "Rename a saved connection",
	Args:  cobra.ExactArgs(2),
	RunE:  runConnectRename,
}

var connectEditCmd = &cobra.Command{
	Use:   "edit <name>",
	Short: "Edit a saved connection",
	Args:  cobra.ExactArgs(1),
	RunE:  runConnectEdit,
}

func init() {
	for _, cmd := range []*cobra.Command{connectAddCmd, connectRemoveCmd, connectDefaultCmd, connectRenameCmd, connectEditCmd} {
		cmd.Flags().Bool("local", false, "use project-local connection store (git repo scoped)")
	}
	connectCmd.AddCommand(connectAddCmd, connectListCmd, connectRemoveCmd,
		connectDefaultCmd, connectRenameCmd, connectEditCmd)
	rootCmd.AddCommand(connectCmd)
}

func connectStore(cmd *cobra.Command) *conn.Store {
	local, _ := cmd.Flags().GetBool("local")
	if local {
		if f := config.ProjectConnectionsFile(); f != "" {
			return conn.NewStore(f)
		}
		// not in a git repo — fall back to global
	}
	return conn.NewStore(config.ConnectionsFile())
}

func prompt(scanner *bufio.Scanner, label, defaultVal string) string {
	if defaultVal != "" {
		fmt.Printf("%s [%s]: ", label, defaultVal)
	} else {
		fmt.Printf("%s: ", label)
	}
	scanner.Scan()
	val := strings.TrimSpace(scanner.Text())
	if val == "" {
		return defaultVal
	}
	return val
}

func runConnectAdd(cmd *cobra.Command, _ []string) error {
	scanner := bufio.NewScanner(os.Stdin)

	name := prompt(scanner, "Name", "")
	if name == "" {
		return &CLIError{Code: ExitGeneral, Message: "name is required"}
	}
	host := prompt(scanner, "Host", "localhost")
	portStr := prompt(scanner, "Port", "5432")
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return &CLIError{Code: ExitGeneral, Message: "invalid port", Err: err}
	}
	user := prompt(scanner, "User", "")
	password := prompt(scanner, "Password", "")
	dbname := prompt(scanner, "Database", "")
	sslmode := prompt(scanner, "SSL Mode", "disable")

	store := connectStore(cmd)
	cfg := conn.ConnectionConfig{
		Name:    name,
		Host:    host,
		Port:    port,
		User:    user,
		DBName:  dbname,
		SSLMode: sslmode,
	}

	if err := store.Add(cfg); err != nil {
		return err
	}

	if password != "" {
		creds := conn.NewCredentialStore(conn.OSKeyring{})
		if err := creds.SetPassword(name, password); err != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to store password in keyring: %v\n", err)
		}
	}

	fmt.Printf("Connection %q saved.\n", name)
	return nil
}

func runConnectList(_ *cobra.Command, _ []string) error {
	type entry struct {
		cfg    conn.ConnectionConfig
		source string
		isDef  bool
	}
	var entries []entry

	// Project-local connections
	if f := config.ProjectConnectionsFile(); f != "" {
		s := conn.NewStore(f)
		conns, _ := s.List()
		defName := s.DefaultName()
		for _, c := range conns {
			entries = append(entries, entry{cfg: c, source: "local", isDef: c.Name == defName})
		}
	}

	// Global connections
	s := conn.NewStore(config.ConnectionsFile())
	conns, err := s.List()
	if err != nil {
		return err
	}
	defName := s.DefaultName()
	for _, c := range conns {
		entries = append(entries, entry{cfg: c, source: "global", isDef: c.Name == defName})
	}

	if len(entries) == 0 {
		fmt.Println("No saved connections.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	if _, err := fmt.Fprintln(w, "NAME\tHOST\tPORT\tUSER\tDATABASE\tDEFAULT\tSCOPE"); err != nil {
		return err
	}
	for _, e := range entries {
		def := ""
		if e.isDef {
			def = "*"
		}
		if _, err := fmt.Fprintf(w, "%s\t%s\t%d\t%s\t%s\t%s\t%s\n",
			e.cfg.Name, e.cfg.Host, e.cfg.Port, e.cfg.User, e.cfg.DBName, def, e.source); err != nil {
			return err
		}
	}
	return w.Flush()
}

func runConnectRemove(cmd *cobra.Command, args []string) error {
	name := args[0]
	store := connectStore(cmd)

	if err := store.Remove(name); err != nil {
		return err
	}

	creds := conn.NewCredentialStore(conn.OSKeyring{})
	_ = creds.DeletePassword(name) // best-effort

	fmt.Printf("Connection %q removed.\n", name)
	return nil
}

func runConnectDefault(cmd *cobra.Command, args []string) error {
	store := connectStore(cmd)
	if err := store.SetDefault(args[0]); err != nil {
		return err
	}
	fmt.Printf("Default connection set to %q.\n", args[0])
	return nil
}

func runConnectRename(cmd *cobra.Command, args []string) error {
	store := connectStore(cmd)
	if err := store.Rename(args[0], args[1]); err != nil {
		return err
	}

	// Move keyring credential to new name
	creds := conn.NewCredentialStore(conn.OSKeyring{})
	if pw, err := creds.GetPassword(args[0]); err == nil {
		_ = creds.SetPassword(args[1], pw)
		_ = creds.DeletePassword(args[0])
	}

	fmt.Printf("Connection %q renamed to %q.\n", args[0], args[1])
	return nil
}

func runConnectEdit(cmd *cobra.Command, args []string) error {
	name := args[0]
	store := connectStore(cmd)

	existing, err := store.Get(name)
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(os.Stdin)
	host := prompt(scanner, "Host", existing.Host)
	portStr := prompt(scanner, "Port", strconv.Itoa(existing.Port))
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return &CLIError{Code: ExitGeneral, Message: "invalid port", Err: err}
	}
	user := prompt(scanner, "User", existing.User)
	password := prompt(scanner, "Password (leave blank to keep)", "")
	dbname := prompt(scanner, "Database", existing.DBName)
	sslmode := prompt(scanner, "SSL Mode", existing.SSLMode)

	cfg := conn.ConnectionConfig{
		Name:    name,
		Host:    host,
		Port:    port,
		User:    user,
		DBName:  dbname,
		SSLMode: sslmode,
	}

	if err := store.Add(cfg); err != nil {
		return err
	}

	if password != "" {
		creds := conn.NewCredentialStore(conn.OSKeyring{})
		if err := creds.SetPassword(name, password); err != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to store password in keyring: %v\n", err)
		}
	}

	fmt.Printf("Connection %q updated.\n", name)
	return nil
}
