package generator

import (
    "os"
    "path/filepath"
    "strings"
    "testing"

    "github.com/hunterkritik-byte/autoline-cli/internal/detector"
)

func TestTemplatesUseBuildKitCaching(t *testing.T) {
    cases := []detector.Stack{
        {Name: "Node.js", PackageManager: "npm", DependencyFiles: []string{"package.json", "package-lock.json"}, BuildCommand: "npm run build", RuntimeCommand: "npm start"},
        {Name: "Python", PackageManager: "pip", DependencyFiles: []string{"requirements.txt"}},
        {Name: "Go"},
        {Name: "Rust", BinaryName: "demo"},
    }
    for _, stack := range cases {
        docker, _, _ := Templates(stack)
        if !strings.Contains(docker, "# syntax=docker/dockerfile:1.7") || !strings.Contains(docker, "--mount=type=cache") {
            t.Errorf("%s template is missing BuildKit caching", stack.Name)
        }
    }
}

func TestGenerateRefusesOverwriteWithoutForce(t *testing.T) {
    root := t.TempDir()
    path := filepath.Join(root, "Dockerfile")
    if err := os.WriteFile(path, []byte("keep"), 0o644); err != nil { t.Fatal(err) }
    stack := detector.Stack{Name: "Generic"}
    if err := Generate(root, stack, false); err == nil { t.Fatal("expected overwrite protection error") }
    if err := Generate(root, stack, true); err != nil { t.Fatal(err) }
    data, err := os.ReadFile(path)
    if err != nil { t.Fatal(err) }
    if string(data) == "keep" { t.Fatal("force did not overwrite Dockerfile") }
}
