package migrate

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	mg "github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/spf13/cobra"

	"github.com/go-kyugo/kygo/internal/config"
	"github.com/go-kyugo/kygo/internal/ui"
)

func Run(migrationsPath, database string, args ...string) error {
	if database == "" {
		return fmt.Errorf("database URL is required (use --database or set DATABASE_URL)")
	}

	ui.Info(fmt.Sprintf("Running migrations from %s -> %s", migrationsPath, database))
	abs, err := filepath.Abs(migrationsPath)
	if err != nil {
		return err
	}
	pathArg := "file://" + abs

	m, err := mg.New(pathArg, database)
	if err != nil {
		ui.Errorf("Failed to initialize migrate: %w", err)
		return err
	}
	defer func() {
		_, _ = m.Close()
	}()

	if len(args) == 0 {
		ui.Errorf("No migrate action provided")
		return err
	}

	action := args[0]
	switch action {
	case "up":
		if len(args) == 1 {
			if err := m.Up(); err != nil && err != mg.ErrNoChange {
				return err
			}
			ui.Success("Migrations completed")
			return nil
		}
		// step value provided
		steps, err := strconv.Atoi(args[1])
		if err != nil {
			ui.Errorf(fmt.Sprintf("Invalid steps: %v", err))
			return err
		}
		if err := m.Steps(steps); err != nil && err != mg.ErrNoChange {
			return err
		}
		ui.Success("Migrations completed")
		return nil

	case "down":
		steps := 1
		if len(args) == 2 {
			s, err := strconv.Atoi(args[1])
			if err != nil {
				return fmt.Errorf("invalid steps: %w", err)
			}
			steps = s
		}
		if err := m.Steps(-steps); err != nil && err != mg.ErrNoChange {
			return err
		}
		ui.Success("Rollback completed")
		return nil

	case "force":
		if len(args) < 2 {
			return fmt.Errorf("force requires a version argument")
		}
		v, err := strconv.Atoi(args[1])
		if err != nil {
			return fmt.Errorf("invalid version: %w", err)
		}
		if err := m.Force(v); err != nil {
			return err
		}
		ui.Success("Force applied")
		return nil

	case "version":
		v, dirty, err := m.Version()
		if err != nil {
			if err == mg.ErrNilVersion {
				ui.Info("No migration version set")
				return nil
			}
			return err
		}
		if dirty {
			ui.Info(fmt.Sprintf("Version: %d (dirty)", v))
		} else {
			ui.Info(fmt.Sprintf("Version: %d", v))
		}
		return nil

	default:
		ui.Errorf("Unknown migrate action: %s", action)
		return err
	}
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
