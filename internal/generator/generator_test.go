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

func TestNodeDockerfileSnapshot(t *testing.T) {
    stack := detector.Stack{
        Name: "Node.js", PackageManager: "npm",
        DependencyFiles: []string{"package.json", "package-lock.json"},
        BuildCommand: "npm run build", RuntimeCommand: "npm start",
    }
    got := nodeDocker(stack)
    expected, err := os.ReadFile(filepath.Join("testdata", "expected_dockerfile.txt"))
    if err != nil { t.Fatal(err) }
    if got != string(expected) {
        t.Fatalf("Dockerfile snapshot changed unexpectedly; update testdata/expected_dockerfile.txt only when the template change is intentional")
    }
}

func TestProvidersGenerateExpectedPipeline(t *testing.T) {
    stack := detector.Stack{Name: "Go"}
    for _, provider := range []Provider{ProviderGitHub, ProviderGitLab, ProviderBitbucket} {
        files := FilesForProvider(stack, provider)
        if len(files) != 3 { t.Fatalf("%s generated %d files, want 3", provider, len(files)) }
        joined := files[1].Content
        switch provider {
        case ProviderGitHub:
            if !strings.Contains(joined, "actions/checkout@v4") { t.Fatal("GitHub workflow missing checkout") }
        case ProviderGitLab:
            if !strings.Contains(joined, "stages:") { t.Fatal("GitLab pipeline missing stages") }
        case ProviderBitbucket:
            if !strings.Contains(joined, "pipelines:") { t.Fatal("Bitbucket pipeline missing pipelines section") }
        }
    }
}

func TestParseProviderRejectsUnknownProvider(t *testing.T) {
    if _, err := ParseProvider("jenkins"); err == nil { t.Fatal("expected unsupported provider error") }
}
