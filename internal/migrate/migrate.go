package migrate

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/go-kyugo/kyugo/internal/config"
	"github.com/go-kyugo/kyugo/internal/ui"
)

func Run(migrationsPath, database string, args ...string) error {
	if database == "" {
		return fmt.Errorf("database URL is required (use --database or set DATABASE_URL)")
	}
	bin, err := exec.LookPath("migrate")
	if err != nil {
		return fmt.Errorf("migrate binary not found in PATH: %w", err)
	}

	ui.Info(fmt.Sprintf("Running migrations from %s -> %s", migrationsPath, database))
	abs, err := filepath.Abs(migrationsPath)
	if err != nil {
		return err
	}
	pathArg := "file://" + abs
	cmdArgs := []string{"-path", pathArg, "-database", database}
	cmdArgs = append(cmdArgs, args...)
	cmd := exec.Command(bin, cmdArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return err
	}
	ui.Success("Migrations completed")
	return nil
}

func makeUpCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "up [steps]",
		Short: "Run up migrations (all or given steps)",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			path, _ := c.Flags().GetString("path")
			database, _ := c.Flags().GetString("database")
			if len(args) == 0 {
				return Run(path, database, "up")
			}
			return Run(path, database, "up", args[0])
		},
	}
	cmd.Flags().String("path", "database/migrations", "migrations directory")
	// default database from config.json if available, else env DATABASE_URL
	dbDefault := os.Getenv("DATABASE_URL")
	if cfg, err := config.Load(""); err == nil {
		if d := cfg.DatabaseURL(); d != "" {
			dbDefault = d
		}
	}
	cmd.Flags().String("database", dbDefault, "database URL")
	return cmd
}

func makeRollbackCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rollback [steps]",
		Short: "Rollback (down) migrations (defaults to 1 step)",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			path, _ := c.Flags().GetString("path")
			database, _ := c.Flags().GetString("database")
			steps := "1"
			if len(args) == 1 {
				steps = args[0]
			}
			return Run(path, database, "down", steps)
		},
	}
	cmd.Flags().String("path", "database/migrations", "migrations directory")
	dbDefault := os.Getenv("DATABASE_URL")
	if cfg, err := config.Load(""); err == nil {
		if d := cfg.DatabaseURL(); d != "" {
			dbDefault = d
		}
	}
	cmd.Flags().String("database", dbDefault, "database URL")
	return cmd
}

func makeForceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "force <version>",
		Short: "Set migration version without running migrations",
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			path, _ := c.Flags().GetString("path")
			database, _ := c.Flags().GetString("database")
			return Run(path, database, "force", args[0])
		},
	}
	cmd.Flags().String("path", "database/migrations", "migrations directory")
	dbDefault := os.Getenv("DATABASE_URL")
	if cfg, err := config.Load(""); err == nil {
		if d := cfg.DatabaseURL(); d != "" {
			dbDefault = d
		}
	}
	cmd.Flags().String("database", dbDefault, "database URL")
	return cmd
}

func makeVersionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print the current migration version and state",
		RunE: func(c *cobra.Command, args []string) error {
			path, _ := c.Flags().GetString("path")
			database, _ := c.Flags().GetString("database")
			return Run(path, database, "version")
		},
	}
	cmd.Flags().String("path", "database/migrations", "migrations directory")
	dbDefault := os.Getenv("DATABASE_URL")
	if cfg, err := config.Load(""); err == nil {
		if d := cfg.DatabaseURL(); d != "" {
			dbDefault = d
		}
	}
	cmd.Flags().String("database", dbDefault, "database URL")
	return cmd
}

func MigrateCmd() *cobra.Command {
	migrateCmd := &cobra.Command{
		Use:   "migrate",
		Short: "Database migration commands",
	}
	migrateCmd.AddCommand(makeUpCmd(), makeRollbackCmd(), makeForceCmd(), makeVersionCmd())
	return migrateCmd
}
