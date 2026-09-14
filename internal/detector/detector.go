package detector

import (
    "encoding/json"
    "fmt"
    "os"
    "path/filepath"
)

type Stack struct {
    Name             string
    Language         string
    PackageManager   string
    BuildCommand     string
    StartCommand     string
    DependencyFiles  []string
    RuntimeCommand   string
    BinaryName       string
}

func Detect(root string) (Stack, error) {
    info, err := os.Stat(root)
    if err != nil {
        return Stack{}, fmt.Errorf("inspect repository: %w", err)
    }
    if !info.IsDir() {
        return Stack{}, fmt.Errorf("repository path is not a directory: %s", root)
    }

    if exists(root, "package.json") {
        stack := Stack{Name: "Node.js", Language: "javascript", PackageManager: detectNodeManager(root), DependencyFiles: nodeDependencyFiles(root), BuildCommand: "npm run build", StartCommand: "npm start", RuntimeCommand: "npm start"}
        if manager := stack.PackageManager; manager == "pnpm" {
            stack.BuildCommand, stack.StartCommand, stack.RuntimeCommand = "pnpm run build", "pnpm start", "pnpm start"
        } else if manager == "yarn" {
            stack.BuildCommand, stack.StartCommand, stack.RuntimeCommand = "yarn build", "yarn start", "yarn start"
        }
        if build, start := nodeScripts(root); build != "" || start != "" {
            if build != "" { stack.BuildCommand = build }
            if start != "" { stack.StartCommand, stack.RuntimeCommand = start, start }
        }
        return stack, nil
    }
    if exists(root, "go.mod") {
        return Stack{Name: "Go", Language: "go", PackageManager: "go", DependencyFiles: []string{"go.mod", "go.sum"}, BuildCommand: "go build -trimpath -ldflags='-s -w' -o /out/autoline-app .", StartCommand: "/autoline-app", RuntimeCommand: "/autoline-app", BinaryName: "autoline-app"}, nil
    }
    if exists(root, "Cargo.toml") {
        name := cargoPackageName(root)
        if name == "" { name = "app" }
        return Stack{Name: "Rust", Language: "rust", PackageManager: "cargo", DependencyFiles: []string{"Cargo.toml", "Cargo.lock"}, BuildCommand: "cargo build --release", StartCommand: "/usr/local/bin/" + name, RuntimeCommand: "/usr/local/bin/" + name, BinaryName: name}, nil
    }
    if exists(root, "requirements.txt") || exists(root, "pyproject.toml") {
        deps := []string{}
        if exists(root, "requirements.txt") { deps = append(deps, "requirements.txt") }
        if exists(root, "pyproject.toml") { deps = append(deps, "pyproject.toml") }
        return Stack{Name: "Python", Language: "python", PackageManager: detectPythonManager(root), DependencyFiles: deps, BuildCommand: "python -m compileall .", StartCommand: "python app.py", RuntimeCommand: "python app.py"}, nil
    }
    return Stack{Name: "Generic", Language: "unknown", PackageManager: "unknown", BuildCommand: "echo 'No build detected'", StartCommand: "echo 'No start command detected'", RuntimeCommand: "sh"}, nil
}

func exists(root, name string) bool {
    _, err := os.Stat(filepath.Join(root, name))
    return err == nil
}

func detectNodeManager(root string) string {
    switch {
    case exists(root, "pnpm-lock.yaml"):
        return "pnpm"
    case exists(root, "yarn.lock"):
        return "yarn"
    default:
        return "npm"
    }
}

func nodeDependencyFiles(root string) []string {
    files := []string{"package.json"}
    for _, name := range []string{"package-lock.json", "pnpm-lock.yaml", "yarn.lock"} {
        if exists(root, name) { files = append(files, name) }
    }
    return files
}

func detectPythonManager(root string) string {
    if exists(root, "uv.lock") { return "uv" }
    if exists(root, "poetry.lock") { return "poetry" }
    return "pip"
}

func nodeScripts(root string) (string, string) {
    data, err := os.ReadFile(filepath.Join(root, "package.json"))
    if err != nil { return "", "" }
    var manifest struct { Scripts map[string]string `json:"scripts"` }
    if err := json.Unmarshal(data, &manifest); err != nil { return "", "" }
    build, start := "", ""
    if _, ok := manifest.Scripts["build"]; ok { build = "npm run build" }
    if _, ok := manifest.Scripts["start"]; ok { start = "npm start" }
    return build, start
}

func cargoPackageName(root string) string {
    data, err := os.ReadFile(filepath.Join(root, "Cargo.toml"))
    if err != nil { return "" }
    inPackage := false
    for _, line := range splitLines(string(data)) {
        line = trimSpace(line)
        if line == "[package]" { inPackage = true; continue }
        if len(line) > 0 && line[0] == '[' { inPackage = false }
        if inPackage && hasPrefix(line, "name") {
            parts := splitEqual(line)
            if len(parts) == 2 { return trimQuotes(trimSpace(parts[1])) }
        }
    }
    return ""
}

func splitLines(s string) []string { return filepathSplitLines(s) }
func filepathSplitLines(s string) []string {
    var lines []string
    start := 0
    for i := 0; i < len(s); i++ {
        if s[i] == '\n' { lines = append(lines, s[start:i]); start = i + 1 }
    }
    if start <= len(s) { lines = append(lines, s[start:]) }
    return lines
}
func trimSpace(s string) string { return stringTrimSpace(s) }
func stringTrimSpace(s string) string { return trimASCII(s) }
func trimASCII(s string) string {
    start, end := 0, len(s)
    for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\r') { start++ }
    for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\r') { end-- }
    return s[start:end]
}
func hasPrefix(s, prefix string) bool {
    if len(s) < len(prefix) { return false }
    return s[:len(prefix)] == prefix
}
func splitEqual(s string) []string {
    for i := 0; i < len(s); i++ { if s[i] == '=' { return []string{s[:i], s[i+1:]} } }
    return nil
}
func trimQuotes(s string) string {
    if len(s) >= 2 && ((s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'')) { return s[1:len(s)-1] }
    return s
}
