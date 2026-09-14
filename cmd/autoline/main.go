package main

import (
    "encoding/json"
    "fmt"
    "os"

    "github.com/fatih/color"
    "github.com/hunterkritik-byte/autoline-cli/internal/detector"
    "github.com/hunterkritik-byte/autoline-cli/internal/generator"
    "github.com/spf13/cobra"
)

const banner = `
    _         _        _     _
   / \  _   _| |_ ___ | |   (_)_ __   ___
  / _ \| | | | __/ _ \| |   | | '_ \ / _ \
 / ___ \ |_| | || (_) | |___| | | | |  __/
/_/   \_\__,_|\__\___/|_____|_|_| |_|\___|
`

var version = "dev"

type scanOutput struct {
    Path           string   `json:"path"`
    Stack          string   `json:"stack"`
    Language       string   `json:"language"`
    PackageManager string   `json:"package_manager"`
    Files          []string `json:"files"`
    DryRun         bool     `json:"dry_run"`
}

func main() {
    var force, dryRun, jsonOutput bool
    root := &cobra.Command{
        Use:           "autoline",
        Short:         "Zero-config repository automation",
        SilenceUsage:  true,
        SilenceErrors: true,
        Version:       version,
        Run: func(cmd *cobra.Command, args []string) {
            fmt.Println(banner)
            _ = cmd.Help()
        },
    }

    scan := &cobra.Command{
        Use:   "scan [path]",
        Args:  cobra.MaximumNArgs(1),
        Short: "Detect the stack and generate CI/CD assets",
        RunE: func(cmd *cobra.Command, args []string) error {
            rootPath := "."
            if len(args) == 1 {
                rootPath = args[0]
            }
            stack, err := detector.Detect(rootPath)
            if err != nil {
                return err
            }
            files := generator.Files(stack)
            names := make([]string, 0, len(files))
            for _, file := range files {
                names = append(names, file.Path)
            }

            if jsonOutput {
                if dryRun {
                    return printJSON(scanOutput{Path: rootPath, Stack: stack.Name, Language: stack.Language, PackageManager: stack.PackageManager, Files: names, DryRun: true})
                }
                if err := generator.Generate(rootPath, stack, force); err != nil {
                    return err
                }
                return printJSON(scanOutput{Path: rootPath, Stack: stack.Name, Language: stack.Language, PackageManager: stack.PackageManager, Files: names, DryRun: false})
            }

            fmt.Println(banner)
            color.Cyan("Scanning: %s", rootPath)
            color.Green("Detected: %s (%s, %s)", stack.Name, stack.Language, stack.PackageManager)
            if dryRun {
                color.Yellow("Dry run — would generate:")
                for _, file := range files { fmt.Printf("  • %s\n", file.Path) }
                return nil
            }
            if err := generator.Generate(rootPath, stack, force); err != nil {
                return err
            }
            color.Green("Generated production-ready Docker, CI, and pre-commit assets.")
            return nil
        },
    }
    scan.Flags().BoolVar(&force, "force", false, "overwrite existing generated files")
    scan.Flags().BoolVar(&dryRun, "dry-run", false, "detect and preview generated files without writing")
    scan.Flags().BoolVar(&jsonOutput, "json", false, "emit machine-readable scan output")
    root.AddCommand(scan)
    root.SetHelpTemplate("AutoLine — repository automation\n\n{{.UsageString}}")

    if err := root.Execute(); err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }
}

func printJSON(out scanOutput) error {
    data, err := json.MarshalIndent(out, "", "  ")
    if err != nil { return err }
    fmt.Println(string(data))
    return nil
}
