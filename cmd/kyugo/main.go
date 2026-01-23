package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"kyugo.dev/kyugo-cli/v1/internal/create"
	initpkg "kyugo.dev/kyugo-cli/v1/internal/init"
	migrate "kyugo.dev/kyugo-cli/v1/internal/migrate"
)

var rootCmd = &cobra.Command{
	Use:   "kyugo",
	Short: "kyugo CLI",
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
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
