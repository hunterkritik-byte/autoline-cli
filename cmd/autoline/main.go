package main

import (
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

func main() {
    var force bool
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
            fmt.Println(banner)
            color.Cyan("Scanning: %s", rootPath)
            stack, err := detector.Detect(rootPath)
            if err != nil {
                return err
            }
            color.Green("Detected: %s (%s, %s)", stack.Name, stack.Language, stack.PackageManager)
            if err := generator.Generate(rootPath, stack, force); err != nil {
                return err
            }
            color.Green("Generated production-ready Docker, CI, and pre-commit assets.")
            return nil
        },
    }
    scan.Flags().BoolVar(&force, "force", false, "overwrite existing generated files")
    root.AddCommand(scan)
    root.SetHelpTemplate("AutoLine — repository automation\n\n{{.UsageString}}")

    if err := root.Execute(); err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }
}
