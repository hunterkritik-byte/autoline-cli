package generator

import (
    "fmt"
    "os"
    "path/filepath"
    "strings"

    "github.com/hunterkritik-byte/autoline-cli/internal/detector"
)

func Generate(root string, stack detector.Stack, force bool) error {
    docker, workflow, precommit := Templates(stack)
    files := map[string]string{"Dockerfile": docker, ".github/workflows/autoline.yml": workflow, ".pre-commit-config.yaml": precommit}
    for name, content := range files {
        path := filepath.Join(root, name)
        if !force {
            if _, err := os.Stat(path); err == nil { return fmt.Errorf("refusing to overwrite %s; rerun with --force", name) }
            if !os.IsNotExist(err) { return fmt.Errorf("inspect %s: %w", name, err) }
        }
        if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil { return fmt.Errorf("create directory for %s: %w", name, err) }
        if err := os.WriteFile(path, []byte(content), 0o644); err != nil { return fmt.Errorf("write %s: %w", name, err) }
    }
    return nil
}

func Templates(s detector.Stack) (string, string, string) {
    switch s.Name {
    case "Node.js": return nodeDocker(s), nodeWorkflow(s), precommit()
    case "Python": return pythonDocker(s), pythonWorkflow(s), precommit()
    case "Go": return goDocker(s), goWorkflow(s), precommit()
    case "Rust": return rustDocker(s), rustWorkflow(s), precommit()
    default: return genericDocker(), genericWorkflow(), precommit()
    }
}

func nodeDocker(s detector.Stack) string {
    install := installCommand(s)
    build := s.BuildCommand
    if s.PackageManager == "pnpm" && strings.HasPrefix(build, "npm run") { build = strings.Replace(build, "npm run", "pnpm run", 1) }
    if s.PackageManager == "yarn" && strings.HasPrefix(build, "npm run") { build = strings.Replace(build, "npm run", "yarn", 1) }
    return fmt.Sprintf(`# syntax=docker/dockerfile:1.7
FROM node:22-bookworm AS build
WORKDIR /app
COPY %s ./
RUN --mount=type=cache,target=/root/.npm %s
COPY . .
RUN %s

FROM node:22-bookworm-slim
WORKDIR /app
ENV NODE_ENV=production
COPY --from=build /app ./
CMD ["%s"]
`, strings.Join(s.DependencyFiles, " "), install, build, shellJSON(s.RuntimeCommand))
}

func pythonDocker(s detector.Stack) string {
    if len(s.DependencyFiles) == 0 { return genericDocker() }
    dep := strings.Join(s.DependencyFiles, " ")
    install := "python -m pip install --prefix=/install -r requirements.txt"
    if s.PackageManager == "poetry" { install = "python -m pip install poetry && poetry install --only main --no-interaction" }
    if s.PackageManager == "uv" { install = "python -m pip install uv && uv sync --frozen --no-dev" }
    return fmt.Sprintf(`# syntax=docker/dockerfile:1.7
FROM python:3.13-slim AS build
WORKDIR /app
COPY %s ./
RUN --mount=type=cache,target=/root/.cache/pip %s
COPY . .

FROM python:3.13-slim
WORKDIR /app
COPY --from=build /usr/local /usr/local
COPY --from=build /install /usr/local
COPY --from=build /app ./
CMD ["python", "app.py"]
`, dep, install)
}

func goDocker(s detector.Stack) string {
    return `# syntax=docker/dockerfile:1.7
FROM golang:1.24-bookworm AS build
WORKDIR /src
COPY go.mod go.sum* ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download
COPY . .
RUN --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/go-build CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/autoline-app .

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/autoline-app /autoline-app
ENTRYPOINT ["/autoline-app"]
`
}

func rustDocker(s detector.Stack) string {
    return fmt.Sprintf(`# syntax=docker/dockerfile:1.7
FROM rust:1.88-bookworm AS build
WORKDIR /app
COPY Cargo.toml Cargo.lock* ./
COPY src ./src
RUN --mount=type=cache,target=/usr/local/cargo/registry --mount=type=cache,target=/app/target cargo build --release

FROM debian:bookworm-slim
COPY --from=build /app/target/release/%s /usr/local/bin/%s
ENTRYPOINT ["/usr/local/bin/%s"]
`, s.BinaryName, s.BinaryName, s.BinaryName)
}

func genericDocker() string { return `# syntax=docker/dockerfile:1.7
FROM alpine:3.22
WORKDIR /app
COPY . .
CMD ["sh"]
` }

func nodeWorkflow(s detector.Stack) string {
    cache := "npm"
    if s.PackageManager == "pnpm" { cache = "pnpm" }
    if s.PackageManager == "yarn" { cache = "yarn" }
    return fmt.Sprintf(`name: AutoLine CI
on:
  push:
  pull_request:
jobs:
  validate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: '22'
          cache: %s
      - run: %s
      - run: %s
      - uses: docker/setup-buildx-action@v3
      - run: docker buildx build --load --tag autoline-app:ci .
`, cache, installCommand(s), s.BuildCommand)
}
func pythonWorkflow(s detector.Stack) string { return fmt.Sprintf(`name: AutoLine CI
on:
  push:
  pull_request:
jobs:
  validate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-python@v5
        with:
          python-version: '3.13'
          cache: pip
      - run: %s
      - run: %s
      - uses: docker/setup-buildx-action@v3
      - run: docker buildx build --load --tag autoline-app:ci .
`, pythonInstallCommand(s), s.BuildCommand) }
func goWorkflow(s detector.Stack) string { return `name: AutoLine CI
on:
  push:
  pull_request:
jobs:
  validate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.24'
          cache: true
      - run: go test ./...
      - run: go vet ./...
      - run: go build ./...
      - uses: docker/setup-buildx-action@v3
      - run: docker buildx build --load --tag autoline-app:ci .
` }
func rustWorkflow(s detector.Stack) string { return `name: AutoLine CI
on:
  push:
  pull_request:
jobs:
  validate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions-rust-lang/setup-rust-toolchain@v1
        with:
          toolchain: stable
      - run: cargo test --locked
      - run: cargo build --release --locked
      - uses: docker/setup-buildx-action@v3
      - run: docker buildx build --load --tag autoline-app:ci .
` }
func genericWorkflow() string { return `name: AutoLine CI
on:
  push:
  pull_request:
jobs:
  validate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: docker/setup-buildx-action@v3
      - run: docker buildx build --load --tag autoline-app:ci .
` }
func precommit() string { return `repos:
  - repo: https://github.com/pre-commit/pre-commit-hooks
    rev: v5.0.0
    hooks:
      - id: trailing-whitespace
      - id: end-of-file-fixer
      - id: check-yaml
      - id: check-json
      - id: check-added-large-files
      - id: detect-private-key
` }
func installCommand(s detector.Stack) string { if s.PackageManager == "pnpm" { return "corepack enable && pnpm install --frozen-lockfile" }; if s.PackageManager == "yarn" { return "corepack enable && yarn install --immutable" }; return "npm ci" }
func pythonInstallCommand(s detector.Stack) string { if s.PackageManager == "poetry" { return "python -m pip install poetry && poetry install --only main --no-interaction" }; if s.PackageManager == "uv" { return "python -m pip install uv && uv sync --frozen --no-dev" }; return "python -m pip install -r requirements.txt" }
func shellJSON(command string) string { return strings.ReplaceAll(command, `"`, `\"`) }
