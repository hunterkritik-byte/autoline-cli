# AutoLine

> ⚡ High-performance, zero-config repository automation for modern engineering teams.

[![Go](https://img.shields.io/badge/Go-1.21%2B-00ADD8?logo=go)](https://go.dev/) [![License](https://img.shields.io/badge/license-MIT-green)](#license)

## 💖 Sponsor This Project

**Help fund faster, cheaper cloud-native delivery.** AutoLine automates optimized multi-stage Docker builds and CI workflows so engineering teams can standardize build caching, reduce duplicated CI work, and make infrastructure easier to maintain. For teams running frequent CI builds, better Docker layer reuse can translate into meaningful reductions in compute and build time—potentially saving thousands annually depending on workload.

If AutoLine saves your team engineering time or CI spend, please consider sponsoring development. Corporate sponsorship directly supports new language detectors, safer generators, benchmarks, and enterprise-ready CI/CD features.

**Sponsorship & partnership inquiries:** hunterkritik@gmail.com

## What it does

AutoLine scans a repository, detects common technology stacks, and generates practical delivery assets in one command:

- Multi-stage Dockerfile tailored to the detected stack
- GitHub Actions CI workflow
- Pre-commit quality hooks
- Node.js, Python, Go, Rust, and generic fallback detection
- Deterministic, idempotent generation

## Quick start

```bash
go install github.com/hunterkritik-byte/autoline-cli/cmd/autoline@latest
cd your-project
autoline scan .
```

Generated files:

```text
Dockerfile
.github/workflows/autoline.yml
.pre-commit-config.yaml
```

## Architecture

```text
cmd/autoline/main.go          CLI entrypoint
internal/detector/            Stack detection
internal/generator/           Docker/CI generation
.github/FUNDING.yml           Sponsorship configuration
```

## Development

```bash
go mod tidy
go test ./...
go vet ./...
go run ./cmd/autoline scan .
```

## Roadmap

- Smarter package-manager and monorepo detection
- BuildKit cache configuration and registry-backed caching
- More framework-specific Docker optimizations
- Additional CI providers
- Snapshot tests for generated assets

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md). Issues and focused pull requests are welcome.

## License

MIT. See `LICENSE` when distributed with a release.
