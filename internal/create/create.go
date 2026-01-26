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

	"github.com/go-kyugo/kygo/internal/ui"
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

		module := "github.com/go-kyugo/kygo"
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
	if err := writeFile(root, dir, filename, buf.Bytes()); err != nil {
		return err
	}

	// If we just created a controller, ensure it's registered in http/route/route.go
	if kind == "controller" {
		routePath := filepath.Join(root, "http", "route", "route.go")
		if _, err := os.Stat(routePath); err == nil {
			b, err := os.ReadFile(routePath)
			if err == nil {
				s := string(b)

				// If the file already references the controller's NewController(), do nothing
				if !strings.Contains(s, data.Name+".NewController()") {
					// Ensure import for the controller package exists
					importPath := module + "/http/controller/" + data.Name
					if !strings.Contains(s, "\""+importPath+"\"") {
						if impIdx := strings.Index(s, "import ("); impIdx != -1 {
							// find end of import block
							rest := s[impIdx:]
							if impEnd := strings.Index(rest, ")"); impEnd != -1 {
								newImportBlock := rest[:impEnd] + "\n\t\"" + importPath + "\"" + rest[impEnd:]
								s = s[:impIdx] + newImportBlock + s[impIdx+len(rest):]
							}
						} else if singleImpIdx := strings.Index(s, "import \""); singleImpIdx != -1 {
							// convert single import to block
							// find end of the import line
							lineEnd := strings.Index(s[singleImpIdx:], "\n")
							if lineEnd != -1 {
								before := s[:singleImpIdx]
								impLine := s[singleImpIdx : singleImpIdx+lineEnd]
								after := s[singleImpIdx+lineEnd:]
								impLine = strings.TrimSpace(impLine)
								impLine = strings.TrimPrefix(impLine, "import ")
								newBlock := "import (\n\t" + strings.Trim(impLine, "\"") + "\n\t\"" + importPath + "\"\n)"
								s = before + newBlock + after
							}
						}
					}

					// Insert router.Controller(...) before the end of Register()
					if regIdx := strings.Index(s, "func Register("); regIdx != -1 {
						// find the opening brace of the function
						if braceIdx := strings.Index(s[regIdx:], "{"); braceIdx != -1 {
							pos := regIdx + braceIdx
							// scan to matching closing brace
							count := 0
							i := pos
							for ; i < len(s); i++ {
								if s[i] == '{' {
									count++
								} else if s[i] == '}' {
									count--
									if count == 0 {
										break
									}
								}
							}
							if i < len(s) {
								line := "\n\trouter.Controller(" + data.Name + ".NewController())\n"
								s = s[:i] + line + s[i:]
							}
						}
					}

					// write back modified route.go
					_ = os.WriteFile(routePath, []byte(s), 0644)
				}
			}
		}
	}

	return nil
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
