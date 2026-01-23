package main

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/go-kyugo/kyugo-cli/internal/create"
	initpkg "github.com/go-kyugo/kyugo-cli/internal/init"
	migrate "github.com/go-kyugo/kyugo-cli/internal/migrate"
	"github.com/go-kyugo/kyugo-cli/internal/ui"
)

var rootCmd = &cobra.Command{
	Use:     "kyugo",
	Short:   "Kyugo CLI",
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
