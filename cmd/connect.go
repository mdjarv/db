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

func init() {
	connectCmd.AddCommand(connectAddCmd, connectListCmd, connectRemoveCmd)
	rootCmd.AddCommand(connectCmd)
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

func runConnectAdd(_ *cobra.Command, _ []string) error {
	scanner := bufio.NewScanner(os.Stdin)

	name := prompt(scanner, "Name", "")
	if name == "" {
		return fmt.Errorf("name is required")
	}
	host := prompt(scanner, "Host", "localhost")
	portStr := prompt(scanner, "Port", "5432")
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return fmt.Errorf("invalid port: %w", err)
	}
	user := prompt(scanner, "User", "")
	password := prompt(scanner, "Password", "")
	dbname := prompt(scanner, "Database", "")
	sslmode := prompt(scanner, "SSL Mode", "disable")

	store := conn.NewStore(config.ConnectionsFile())
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
	store := conn.NewStore(config.ConnectionsFile())
	conns, err := store.List()
	if err != nil {
		return err
	}
	if len(conns) == 0 {
		fmt.Println("No saved connections.")
		return nil
	}

	defName := store.DefaultName()
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	if _, err := fmt.Fprintln(w, "NAME\tHOST\tPORT\tUSER\tDATABASE\tDEFAULT"); err != nil {
		return err
	}
	for _, c := range conns {
		def := ""
		if c.Name == defName {
			def = "*"
		}
		if _, err := fmt.Fprintf(w, "%s\t%s\t%d\t%s\t%s\t%s\n", c.Name, c.Host, c.Port, c.User, c.DBName, def); err != nil {
			return err
		}
	}
	return w.Flush()
}

func runConnectRemove(_ *cobra.Command, args []string) error {
	name := args[0]
	store := conn.NewStore(config.ConnectionsFile())

	if err := store.Remove(name); err != nil {
		return err
	}

	creds := conn.NewCredentialStore(conn.OSKeyring{})
	_ = creds.DeletePassword(name) // best-effort

	fmt.Printf("Connection %q removed.\n", name)
	return nil
}
