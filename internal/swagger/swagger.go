package swagger

import (
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"

	"github.com/go-kyugo/kygo/internal/ui"
)

// SwaggerCmd returns a cobra command with subcommands for swagger tooling.
func SwaggerCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "swagger",
		Short: "Swagger generation and tooling",
	}

	// helper to run swag via `go run`
	runSwag := func(dir string, args []string) error {
		if _, err := exec.LookPath("go"); err != nil {
			ui.Errorf("Go tool not found in PATH: %w", err)
			return err
		}

		c := exec.Command("go", args...)
		c.Dir = dir
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		c.Stdin = os.Stdin

		ui.Info("Running: go " + strings.Join(args, " "))
		if err := c.Run(); err != nil {
			ui.Errorf("Failed to run swag via go run: %w", err)
			return err
		}
		return nil
	}

	// init subcommand: generate docs (alias behavior)
	var dir string
	var mainFile string
	var outDir string
	initCmd := &cobra.Command{
		Use:   "init",
		Short: "Generate swagger docs from annotations (runs swag via go run)",
		RunE: func(cmd *cobra.Command, args []string) error {
			runArgs := []string{"run", "github.com/swaggo/swag/cmd/swag@latest", "init"}
			if mainFile != "" {
				runArgs = append(runArgs, "-g", mainFile)
			}
			if outDir != "" {
				runArgs = append(runArgs, "-o", outDir)
			}

			if err := runSwag(dir, runArgs); err != nil {
				ui.Errorf("%v", err)
				return err
			}

			ui.Successf("Swagger generated in: %s", outDir)
			return nil
		},
	}

	initCmd.Flags().StringVarP(&dir, "dir", "C", ".", "Project Directory (default: .)")
	initCmd.Flags().StringVarP(&mainFile, "main", "g", "main.go", "Main file to analyze (default: main.go)")
	initCmd.Flags().StringVarP(&outDir, "out", "o", "resources/docs", "Output directory for Swagger docs (default: resources/docs)")

	// generate subcommand: alias to init (for semantics)
	var dirG string
	var mainFileG string
	var outDirG string
	generateCmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate swagger docs (alias to init)",
		RunE: func(cmd *cobra.Command, args []string) error {
			runArgs := []string{"run", "github.com/swaggo/swag/cmd/swag@latest", "init"}
			if mainFileG != "" {
				runArgs = append(runArgs, "-g", mainFileG)
			}
			if outDirG != "" {
				runArgs = append(runArgs, "-o", outDirG)
			}

			if err := runSwag(dirG, runArgs); err != nil {
				ui.Errorf("%v", err)
				return err
			}

			ui.Successf("Swagger generated in: %s", outDirG)
			return nil
		},
	}

	generateCmd.Flags().StringVarP(&dirG, "dir", "C", ".", "Project Directory (default: .)")
	generateCmd.Flags().StringVarP(&mainFileG, "main", "g", "main.go", "Main file to analyze (default: main.go)")
	generateCmd.Flags().StringVarP(&outDirG, "out", "o", "resources/docs", "Output directory for Swagger docs (default: resources/docs)")

	root.AddCommand(initCmd, generateCmd)
	return root
}
