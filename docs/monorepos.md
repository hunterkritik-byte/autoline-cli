# Monorepos and Polyglot Workspaces

AutoLine can discover independent services nested below a repository root. A workspace such as:

```text
frontend/package.json
backend/go.mod
worker/pyproject.toml
```

is represented as three modules rather than being reduced to one root stack.

## Detection rules

AutoLine recursively walks the repository and skips dependency/build directories such as `node_modules`, `vendor`, `target`, virtual environments, and VCS metadata. Supported application manifests are treated as module signals.

Modules are sorted by relative path so JSON output and generated plans remain deterministic.

## Generation model

Each module receives its own Dockerfile, pre-commit configuration, and selected CI provider configuration. This allows teams to evolve services independently while retaining one automation contract.

Use:

```bash
autoline scan . --provider=github
autoline scan . --provider=gitlab --dry-run
autoline scan . --provider=bitbucket --json
```

## Review considerations

Monorepos frequently contain shared workspaces, custom build orchestration, and cross-service dependencies. Review generated paths and commands before deployment, especially for Node.js workspaces, Cargo workspaces, Go projects with `cmd/` entrypoints, and Python applications whose runtime entrypoint is not `app.py`.
