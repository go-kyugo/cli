package create

import (
	"bytes"
	"embed"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/spf13/cobra"

	"kyugo.dev/kyugo-cli/v1/internal/ui"
)

//go:embed templates/*.gotmpl
var templatesFS embed.FS

var tmpl *template.Template

var CreateCmd = &cobra.Command{
	Use:     "create <type> <name>",
	Short:   "Create project artefacts",
	Aliases: []string{"g"},
	Args:    cobra.MaximumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {

		if len(args) == 0 {
			_ = cmd.Help()
			ui.Println()
			ui.Usage("usage: kyugo create <type> <name>")
			return nil
		}

		if len(args) < 2 {
			return nil
		}

		kind := args[0]
		name := args[1]

		cwd, err := os.Getwd()
		if err != nil {
			return err
		}

		module := "kyugo.dev/kyugo-cli/v1"
		root := cwd
		modPath := filepath.Join(root, "go.mod")
		if _, err := os.Stat(modPath); os.IsNotExist(err) {
			root = filepath.Dir(root)
		}

		if err := Generate(root, module, kind, name); err != nil {
			return err
		}
		ui.Successf("Created %s %s", kind, name)
		return nil
	},
}

func CreateKindCmd(kind string) *cobra.Command {
	return &cobra.Command{
		Use:   kind + " <name>",
		Short: "Create " + kind + " artefact",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			cwd, err := os.Getwd()
			if err != nil {
				return err
			}

			module := "kyugo.dev/kyugo-cli/v1"
			root := cwd
			modPath := filepath.Join(root, "go.mod")
			if _, err := os.Stat(modPath); os.IsNotExist(err) {
				root = filepath.Dir(root)
			}

			if err := Generate(root, module, kind, name); err != nil {
				return err
			}
			ui.Successf("Created %s %s", kind, name)
			return nil
		},
	}
}

func init() {
	tmpl = template.Must(template.ParseFS(templatesFS, "templates/*.gotmpl"))
}

func Generate(root, module, kind, name string) error {
	n := sanitizeName(name)
	data := struct {
		Name       string
		StructName string
		FuncName   string
		ModelName  string
		Table      string
	}{
		Name:       n,
		StructName: toPascal(n),
		FuncName:   toLowerFirst(toPascal(n) + "Controller"),
		ModelName:  toPascal(n),
		Table:      toSnake(n),
	}

	var tplName string
	var filename string
	switch kind {
	case "controller":
		tplName = "controller.gotmpl"
		filename = n + ".go"
	case "model":
		tplName = "model.gotmpl"
		filename = n + ".go"
	case "repository":
		tplName = "repository.gotmpl"
		filename = n + ".go"
	case "service":
		tplName = "service.gotmpl"
		filename = n + ".go"
	case "middleware":
		tplName = "middleware.gotmpl"
		filename = n + ".go"
	case "migration":
		tplName = "migration.gotmpl"
		ts := time.Now().Format("20060102150405")
		// use .up.sql / .down.sql suffixes to be compatible with golang-migrate
		upFilename := fmt.Sprintf("%s_%s.up.sql", ts, toSnake(n))
		downFilename := fmt.Sprintf("%s_%s.down.sql", ts, toSnake(n))

		var upBuf bytes.Buffer
		if err := tmpl.ExecuteTemplate(&upBuf, tplName, data); err != nil {
			return err
		}

		// prefer an embedded migration_down.gotmpl if present
		var downBuf bytes.Buffer
		if t := tmpl.Lookup("migration_down.gotmpl"); t != nil {
			if err := t.Execute(&downBuf, data); err != nil {
				return err
			}
		} else {
			// default: leave a TODO for manual down migration since operations vary (add column, change, etc.)
			downTmplText := "-- TODO: implement down migration for {{ .Table }}\n-- This migration was generated automatically; edit as needed.\n"
			downTmpl, err := template.New("migration_down_default").Parse(downTmplText)
			if err != nil {
				return err
			}
			if err := downTmpl.Execute(&downBuf, data); err != nil {
				return err
			}
		}

		dir := kindDir(kind, n)
		if err := writeFile(root, dir, upFilename, upBuf.Bytes()); err != nil {
			return err
		}
		if err := writeFile(root, dir, downFilename, downBuf.Bytes()); err != nil {
			return err
		}
		return nil
	case "seed":
		tplName = "seed.gotmpl"
		filename = n + ".go"
	case "dto":
		tplName = "dto.gotmpl"
		filename = n + ".go"
	case "validation":
		tplName = "validation.gotmpl"
		filename = n + ".go"
	default:
		return errors.New("unknown generate type: " + kind)
	}

	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, tplName, data); err != nil {
		return err
	}

	dir := kindDir(kind, n)
	return writeFile(root, dir, filename, buf.Bytes())
}

func sanitizeName(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, "-", "_")
	return strings.ToLower(s)
}

func toPascal(s string) string {
	parts := strings.FieldsFunc(s, func(r rune) bool { return r == '_' || r == '-' || r == ' ' })
	for i := range parts {
		if len(parts[i]) == 0 {
			continue
		}
		parts[i] = strings.ToUpper(parts[i][:1]) + strings.ToLower(parts[i][1:])
	}
	return strings.Join(parts, "")
}

func toSnake(s string) string {
	s = strings.ReplaceAll(s, "-", "_")
	var out []rune
	for i, r := range s {
		if r >= 'A' && r <= 'Z' {
			if i > 0 {
				out = append(out, '_')
			}
			out = append(out, r+'a'-'A')
		} else {
			out = append(out, r)
		}
	}
	return string(out)
}

func ensureDir(root, dir string) (string, error) {
	p := filepath.Join(root, dir)
	if err := os.MkdirAll(p, 0755); err != nil {
		return "", err
	}
	return p, nil
}

func writeFile(root, dir, filename string, content []byte) error {
	p, err := ensureDir(root, dir)
	if err != nil {
		return err
	}
	fp := filepath.Join(p, filename)
	if _, err := os.Stat(fp); err == nil {
		return fmt.Errorf("file already exists: %s", fp)
	}
	return os.WriteFile(fp, content, 0644)
}

func toLowerFirst(s string) string {
	if s == "" {
		return s
	}
	return strings.ToLower(s[:1]) + s[1:]
}

func kindDir(kind, name string) string {
	switch kind {
	case "controller":
		return filepath.Join("http", "controller", name)
	case "model":
		return filepath.Join("database", "model")
	case "repository":
		return filepath.Join("database", "repository")
	case "service":
		return filepath.Join("services", name)
	case "middleware":
		return filepath.Join("http", "middleware")
	case "migration":
		return filepath.Join("database", "migrations")
	case "seed":
		return filepath.Join("database", "seed")
	case "dto":
		return "dto"
	case "validation":
		return filepath.Join("http", "validation")
	default:
		return ""
	}
}
