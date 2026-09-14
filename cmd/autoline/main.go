package main

import (
    "fmt"
    "os"

    "github.com/fatih/color"
    "github.com/hunterkritik-byte/autoline-cli/internal/detector"
    "github.com/hunterkritik-byte/autoline-cli/internal/generator"
    "github.com/spf13/cobra"
)

func main() {
    root := &cobra.Command{Use: "autoline", Short: "Zero-config repository automation"}
    root.AddCommand(&cobra.Command{
        Use: "scan [path]", Args: cobra.MaximumNArgs(1), Short: "Detect the stack and generate CI/CD assets",
        RunE: func(cmd *cobra.Command, args []string) error {
            rootPath := "."
            if len(args) == 1 { rootPath = args[0] }
            color.Cyan("AutoLine — scanning %s", rootPath)
            stack, err := detector.Detect(rootPath)
            if err != nil { return err }
            color.Green("Detected: %s", stack.Name)
            if err := generator.Generate(rootPath, stack); err != nil { return err }
            color.Green("Generated Dockerfile, GitHub Actions workflow, and pre-commit configuration.")
            return nil
        },
    })
    root.SetHelpTemplate("AutoLine — repository automation\n\n{{.UsageString}}")
    if err := root.Execute(); err != nil { fmt.Fprintln(os.Stderr, err); os.Exit(1) }
}
