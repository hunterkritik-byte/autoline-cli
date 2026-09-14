# AutoLine Usage

## Installation

AutoLine is a Go CLI. The recommended installation is:

```bash
go install github.com/hunterkritik-byte/autoline-cli/cmd/autoline@latest
```

Go installs the executable under `$(go env GOPATH)/bin` unless `GOBIN` is configured. If the command is installed but your shell reports `autoline: command not found`, add the Go bin directory to your PATH:

```bash
export PATH="$(go env GOPATH)/bin:$PATH"
autoline --help
```

To make that permanent in Bash:

```bash
echo 'export PATH="$(go env GOPATH)/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc
```

For Zsh, add the same line to `~/.zshrc`.

Termux users can use the same process. Verify the executable with:

```bash
ls "$(go env GOPATH)/bin/autoline"
command -v autoline
autoline --help
```

If the first command finds the binary but `command -v` returns nothing, this is a PATH configuration issue, not a failed installation.

## Verify your environment

Run:

```bash
autoline doctor
autoline doctor --json
```

`doctor` checks the local toolchain and helps identify missing tools before generated delivery assets are used.

## First scan: preview first

From the root of your project:

```bash
cd your-project
autoline scan . --dry-run
```

For automation or debugging, use JSON:

```bash
autoline scan . --dry-run --json
```

Review the detected modules and proposed output. When the plan is correct, generate the assets:

```bash
autoline scan .
```

AutoLine does not execute your project's build, test, or start commands while scanning.

## Scan a specific directory

```bash
autoline scan ./services/api --dry-run
```

AutoLine recursively detects independent modules in supported workspaces, which is useful for monorepos and polyglot repositories.

## Choose a CI provider

GitHub Actions is the default provider. You can explicitly select:

```bash
autoline scan . --provider=github
autoline scan . --provider=gitlab
autoline scan . --provider=bitbucket
```

## Inspect supported ecosystems

```bash
autoline languages
autoline languages --json
```

The inventory includes detection support across Node.js, Python, Go, Rust, Java, Kotlin, .NET, PHP, Ruby, Elixir, Dart/Flutter, Swift, C/C++, Scala, Haskell, Lua, Julia, Zig, and Deno.

## Repository intelligence

`autoline insights` is an optional Python-powered specialist command. It requires Python 3 but does not make Python a dependency for the normal Go CLI workflow.

```bash
python3 --version
autoline insights .
autoline insights . --json
```

The insights engine is read-only, uses Python's standard library, skips common generated/dependency directories, never executes project code, and reports credential-pattern counts without printing secret values.

## Replace generated assets

```bash
autoline scan . --force
```

Without `--force`, AutoLine refuses to overwrite an existing generated file. This protects hand-edited Dockerfiles and workflows.

## Generated assets

Depending on the provider and detected stack, AutoLine can generate:

- `Dockerfile`: stack-aware multi-stage image build
- `.github/workflows/autoline.yml`: language-aware validation and Docker build
- `.gitlab-ci.yml`: GitLab CI delivery workflow
- `bitbucket-pipelines.yml`: Bitbucket Pipelines workflow
- `.pre-commit-config.yaml`: baseline repository hygiene hooks

Review generated files before production deployment, particularly for monorepos, custom application entrypoints, native dependencies, secrets, registries, and organization-specific CI policies.

## Recommended user workflow

```text
1. Install AutoLine
2. Ensure $(go env GOPATH)/bin is on PATH
3. Run autoline doctor
4. Preview with autoline scan . --dry-run --json
5. Review detected modules and generated plan
6. Run autoline scan .
7. Review generated Docker/CI files
8. Commit and validate in your own CI
```
