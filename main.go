package main

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/go-kyugo/kygo/internal/create"
	initpkg "github.com/go-kyugo/kygo/internal/init"
	migrate "github.com/go-kyugo/kygo/internal/migrate"
	"github.com/go-kyugo/kygo/internal/ui"
)

var rootCmd = &cobra.Command{
	Use:     "kygo",
	Short:   "Kygo CLI",
	Version: "1.0.0",
}

func init() {
	rootCmd.AddCommand(create.CreateCmd)
	rootCmd.AddCommand(initpkg.MakeInitCmd())
	kinds := []string{"controller", "model", "repository", "service", "middleware", "migration", "seed", "dto", "validation"}
	for _, k := range kinds {
		create.CreateCmd.AddCommand(create.CreateKindCmd(k))
	}
	rootCmd.AddCommand(migrate.MigrateCmd())
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		ui.Errorf("%v", err)
		os.Exit(1)
	}
}
