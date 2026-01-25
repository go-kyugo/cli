package initpkg

import (
	"bytes"
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/spf13/cobra"

	"github.com/go-kyugo/kygo/internal/ui"
)

//go:embed templates/project/**
var projectFS embed.FS

// MakeInitCmd returns the `init` command with `project` subcommand.
func MakeInitCmd() *cobra.Command {

	// remove any placeholder KEEP files copied from templates
	// (these are used only so the embed includes empty dirs)
	removeKEEPs := func(root string) error {
		return filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}
			if d.Name() == "KEEP" {
				if err := os.Remove(p); err != nil {
					return fmt.Errorf("removing KEEP %s: %w", p, err)
				}
			}
			return nil
		})
	}

	projectCmd := &cobra.Command{
		Use:   "init <name>",
		Short: "Create a new project skeleton",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			outDir := name
			if err := os.MkdirAll(outDir, 0755); err != nil {
				return err
			}

			// create standard directories
			dirs := []string{
				"http/controller",
				"http/middleware",
				"http/validation",
				"http/route",
				"dto",
				"service",
				"resources/lang/en-US",
				"database/model",
				"database/repository",
				"database/migrations",
				"database/seed",
			}
			for _, d := range dirs {
				p := filepath.Join(outDir, d)
				if err := os.MkdirAll(p, 0755); err != nil {
					return err
				}
				// add a .gitkeep to ensure directory is tracked
				gitkeep := filepath.Join(p, ".gitkeep")
				if _, err := os.Stat(gitkeep); os.IsNotExist(err) {
					if err := os.WriteFile(gitkeep, []byte(""), 0644); err != nil {
						return err
					}
				}
			}

			// walk embedded templates/project recursively and copy files
			data := struct{ Name string }{Name: name}
			walkRoot := "templates/project"
			if err := fs.WalkDir(projectFS, walkRoot, func(path string, d fs.DirEntry, err error) error {
				if err != nil {
					return err
				}
				// compute relative path under templates/project
				rel := strings.TrimPrefix(path, walkRoot)
				rel = strings.TrimPrefix(rel, string(filepath.Separator))
				if rel == "" {
					return nil
				}
				targetPath := filepath.Join(outDir, rel)
				if d.IsDir() {
					return os.MkdirAll(targetPath, 0755)
				}
				// skip placeholder files used only for embedding
				if d.Name() == "KEEP" {
					return nil
				}
				// read file content from embedded FS
				content, err := projectFS.ReadFile(path)
				if err != nil {
					return err
				}
				// process as template
				tpl, err := template.New(rel).Parse(string(content))
				if err != nil {
					return err
				}
				var buf bytes.Buffer
				if err := tpl.Execute(&buf, data); err != nil {
					return err
				}
				// strip .gotmpl suffix if present
				outName := filepath.Base(rel)
				if filepath.Ext(outName) == ".gotmpl" {
					outName = outName[:len(outName)-len(".gotmpl")]
				}
				outDirPath := filepath.Join(outDir, filepath.Dir(rel))
				if err := os.MkdirAll(outDirPath, 0755); err != nil {
					return err
				}
				outPath := filepath.Join(outDirPath, outName)
				if err := os.WriteFile(outPath, buf.Bytes(), 0644); err != nil {
					return err
				}

				return nil
			}); err != nil {
				return err
			}

			// remove placeholder KEEP files from the generated project
			if err := removeKEEPs(outDir); err != nil {
				return err
			}
			ui.Successf("Created project in %s", outDir)
			return nil
		},
	}

	return projectCmd
}
