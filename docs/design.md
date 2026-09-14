# AutoLine Design

AutoLine separates repository detection from asset generation so new stacks can be added without changing the CLI surface.

## Flow

```text
repository -> detector -> Stack model -> generator -> Dockerfile + CI + pre-commit
```

## Detection

Detection is deterministic and based on repository manifests and lockfiles. Package-manager signals take precedence over generic defaults. The detector returns the commands and dependency files needed by generators.

## Generation

Generators use multi-stage Docker builds and BuildKit cache mounts where practical. Dependency manifests are copied before source files so dependency layers can be reused when application source changes.

Generated assets are intentionally ordinary text files. Teams can review and modify them before committing or building images.

## Safety model

AutoLine does not execute generated commands during `scan`. It only inspects repository metadata and writes files. Generated Dockerfiles and workflows should still be reviewed for project-specific requirements, especially custom build outputs, native dependencies, monorepos, and deployment credentials.

## Extension points

Add a new stack by extending `detector.Stack`, implementing a detector branch, adding Docker/CI templates, and covering the behavior with table-driven tests.
