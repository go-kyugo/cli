# kygo CLI

Installation

To install the binary globally (Go 1.17+):

```bash
go install github.com/go-kyugo/kygo@latest
```

To install a specific version:

```bash
go install github.com/go-kyugo/kygo@v1.0.0
```

If the binary is not available immediately, check where Go places binaries and add it to your `PATH`:

```bash
go env GOBIN
go env GOPATH
export GOBIN="$HOME/go/bin"
export PATH="$PATH:$GOBIN"
```

Usage

After installing, run the `kygo` command (or use `./kygo` during development):

```bash
kygo create <type> <name>
```

Supported types: `controller`, `model`, `repository`, `service`, `middleware`, `migration`, `seed`, `dto`, `validation`.

Examples

```bash
kygo create model user
```

This will create `database/model/user.go` containing a stub `struct`.

More

See the documentation on pkg.go.dev: https://pkg.go.dev/github.com/go-kyugo/kygo

Contributing

Create pull requests or open issues in the repository for improvements or fixes.

Commands

The CLI exposes the following commands (use `kygo --help` for details):

- `create <type> <name>` (alias: `g`): create project artefacts.
	- Supported types: `controller`, `model`, `repository`, `service`, `middleware`, `migration`, `seed`, `dto`, `validation`.
	- Example: `kygo create model user` — creates `database/model/user.go`.

- `init <name>`: create a new project skeleton.
	- Example: `kygo init myapp` — generates a new project directory `myapp` with standard layout and templates.

- `migrate <subcommand>`: database migration commands. All migrate subcommands accept `--path` (default `database/migrations`) and `--database` (default from config.json or `DATABASE_URL`).
	- `migrate up [steps]`: run up migrations (all or given steps). Example: `kygo migrate up` or `kygo migrate up 2`.
	- `migrate rollback [steps]`: rollback (down) migrations (defaults to 1 step). Example: `kygo migrate rollback`.
	- `migrate force <version>`: set migration version without running migrations. Example: `kygo migrate force 20230101120000`.
	- `migrate version`: print current migration version and state.


