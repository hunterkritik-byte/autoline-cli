package main

import (
    "encoding/json"
    "fmt"
    "os"

    "github.com/fatih/color"
    "github.com/hunterkritik-byte/autoline-cli/internal/detector"
    "github.com/hunterkritik-byte/autoline-cli/internal/doctor"
    "github.com/hunterkritik-byte/autoline-cli/internal/generator"
    "github.com/spf13/cobra"
)

const banner = `
    _         _        _     _
   / \\  _   _| |_ ___ | |   (_)_ __   ___
  / _ \\| | | | __/ _ \\| |   | | '_ \\ / _ \\
 / ___ \\ |_| | || (_) | |___| | | | |  __/
/_/   \\_\\__,_|\\__\\___/|_____|_|_| |_|\\___|
`

var version = "dev"

type scanOutput struct {
    Path string `json:"path"`
    Stack string `json:"stack,omitempty"`
    Language string `json:"language,omitempty"`
    PackageManager string `json:"package_manager,omitempty"`
    Provider string `json:"provider"`
    Files []string `json:"files,omitempty"`
    Modules []moduleOutput `json:"modules,omitempty"`
    DryRun bool `json:"dry_run"`
}

type moduleOutput struct {
    Path string `json:"path"`
    Stack string `json:"stack"`
    Language string `json:"language"`
    PackageManager string `json:"package_manager"`
}

type languageInfo struct {
    Name string `json:"name"`
    Language string `json:"language"`
    Manifest string `json:"manifest"`
    PackageManager string `json:"package_manager"`
}

func main() {
    var force, dryRun, jsonOutput bool
    var providerName string
    root := &cobra.Command{
        Use: "autoline", Short: "Zero-config repository automation", SilenceUsage: true, SilenceErrors: true, Version: version,
        Run: func(cmd *cobra.Command, args []string) { fmt.Println(banner); _ = cmd.Help() },
    }

    scan := &cobra.Command{
        Use: "scan [path]", Args: cobra.MaximumNArgs(1), Short: "Detect stacks and generate CI/CD assets",
        RunE: func(cmd *cobra.Command, args []string) error {
            rootPath := "."
            if len(args) == 1 { rootPath = args[0] }
            provider, err := generator.ParseProvider(providerName)
            if err != nil { return err }
            workspace, err := detector.DetectWorkspace(rootPath)
            if err != nil { return err }
            if len(workspace.Modules) == 0 { return fmt.Errorf("no supported application manifests found in %s", rootPath) }

            output := scanOutput{Path: rootPath, Provider: string(provider), DryRun: dryRun}
            if len(workspace.Modules) == 1 {
                stack := workspace.Modules[0].Stack
                output.Stack, output.Language, output.PackageManager = stack.Name, stack.Language, stack.PackageManager
                files := generator.FilesForProvider(stack, provider)
                for _, file := range files { output.Files = append(output.Files, file.Path) }
            } else {
                for _, module := range workspace.Modules { output.Modules = append(output.Modules, moduleOutput{module.Path, module.Stack.Name, module.Stack.Language, module.Stack.PackageManager}) }
            }
            if jsonOutput {
                if !dryRun {
                    if err := generator.GenerateWorkspace(rootPath, workspace, provider, force); err != nil { return err }
                }
                return printJSON(output)
            }

            fmt.Println(banner)
            color.Cyan("Scanning: %s", rootPath)
            color.Green("CI provider: %s", provider)
            if len(workspace.Modules) == 1 {
                color.Green("Detected: %s (%s, %s)", output.Stack, output.Language, output.PackageManager)
                if dryRun { color.Yellow("Dry run — no files written"); return nil }
            } else {
                color.Green("Detected polyglot workspace with %d modules:", len(workspace.Modules))
                for _, module := range workspace.Modules { fmt.Printf("  • %s — %s (%s)\n", module.Path, module.Stack.Name, module.Stack.PackageManager) }
                if dryRun { color.Yellow("Dry run — no files written"); return nil }
            }
            if err := generator.GenerateWorkspace(rootPath, workspace, provider, force); err != nil { return err }
            color.Green("Generated production-ready Docker, CI, pre-commit, OCI metadata, and SBOM automation.")
            return nil
        },
    }
    scan.Flags().BoolVar(&force, "force", false, "overwrite existing generated files")
    scan.Flags().BoolVar(&dryRun, "dry-run", false, "detect and preview generated files without writing")
    scan.Flags().BoolVar(&jsonOutput, "json", false, "emit machine-readable scan output")
    scan.Flags().StringVar(&providerName, "provider", "github", "CI provider: github, gitlab, or bitbucket")

    doctorCmd := &cobra.Command{
        Use: "doctor", Short: "Check local tools required by generated workflows",
        RunE: func(cmd *cobra.Command, args []string) error {
            checks := doctor.Run()
            if jsonOutput {
                if err := printJSON(checks); err != nil { return err }
            } else {
                fmt.Println(banner)
                for _, check := range checks {
                    status := "✗"; if check.Found { status = "✓" }
                    fmt.Printf("%s %-16s %s\n", status, check.Name, check.Message)
                }
            }
            if !doctor.Healthy(checks) { return fmt.Errorf("one or more required tools are unavailable") }
            return nil
        },
    }
    doctorCmd.Flags().BoolVar(&jsonOutput, "json", false, "emit machine-readable diagnostic output")

    languagesCmd := &cobra.Command{
        Use: "languages", Short: "List supported language and toolchain detection",
        RunE: func(cmd *cobra.Command, args []string) error {
            languages := []languageInfo{
                {"Node.js", "javascript", "package.json", "npm / pnpm / yarn / bun"},
                {"Python", "python", "requirements.txt / pyproject.toml", "pip / poetry / uv"},
                {"Go", "go", "go.mod", "go"},
                {"Rust", "rust", "Cargo.toml", "cargo"},
                {"Java", "java", "pom.xml", "maven"},
                {"Kotlin", "kotlin", "build.gradle.kts / settings.gradle.kts", "gradle"},
                {"C#/.NET", "csharp", "*.csproj / *.sln", "dotnet"},
                {"PHP", "php", "composer.json", "composer"},
                {"Ruby", "ruby", "Gemfile", "bundler"},
                {"Elixir", "elixir", "mix.exs", "mix"},
                {"Dart/Flutter", "dart", "pubspec.yaml", "pub"},
                {"Swift", "swift", "Package.swift", "swiftpm"},
                {"C/C++", "cpp", "CMakeLists.txt", "cmake"},
            }
            if jsonOutput { return printJSON(languages) }
            fmt.Println(banner)
            for _, item := range languages { fmt.Printf("%-14s %-10s %-34s %s\n", item.Name, item.Language, item.Manifest, item.PackageManager) }
            return nil
        },
    }
    languagesCmd.Flags().BoolVar(&jsonOutput, "json", false, "emit machine-readable language inventory")

    root.AddCommand(scan, doctorCmd, languagesCmd)
    root.SetHelpTemplate("AutoLine — repository automation\n\n{{.UsageString}}")
    if err := root.Execute(); err != nil { fmt.Fprintln(os.Stderr, err); os.Exit(1) }
}

func printJSON(out any) error {
    data, err := json.MarshalIndent(out, "", "  ")
    if err != nil { return err }
    fmt.Println(string(data))
    return nil
}
