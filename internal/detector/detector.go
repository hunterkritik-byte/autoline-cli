package detector

import (
    "encoding/json"
    "fmt"
    "os"
    "path/filepath"
    "sort"
    "strings"
)

type Stack struct {
    Name            string
    Language        string
    PackageManager  string
    BuildCommand    string
    StartCommand    string
    DependencyFiles []string
    RuntimeCommand  string
    BinaryName      string
}

type Module struct {
    Path  string
    Stack Stack
}

type Workspace struct {
    Root    string
    Modules []Module
}

func Detect(root string) (Stack, error) {
    workspace, err := DetectWorkspace(root)
    if err != nil { return Stack{}, err }
    if len(workspace.Modules) == 0 {
        return Stack{Name: "Generic", Language: "unknown", PackageManager: "unknown", BuildCommand: "echo 'No build detected'", StartCommand: "echo 'No start command detected'", RuntimeCommand: "sh"}, nil
    }
    return workspace.Modules[0].Stack, nil
}

// DetectWorkspace discovers independent application manifests below root.
// Nested dependency/build directories are skipped so dependency trees do not
// turn into dozens of false modules.
func DetectWorkspace(root string) (Workspace, error) {
    info, err := os.Stat(root)
    if err != nil { return Workspace{}, fmt.Errorf("inspect repository: %w", err) }
    if !info.IsDir() { return Workspace{}, fmt.Errorf("repository path is not a directory: %s", root) }

    modules := make([]Module, 0, 8)
    err = filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
        if walkErr != nil { return fmt.Errorf("walk %s: %w", path, walkErr) }
        if entry.IsDir() {
            if path != root && isIgnoredDirectory(entry.Name()) { return filepath.SkipDir }
            return nil
        }
        if !isManifest(filepath.Base(path)) { return nil }
        moduleRoot := filepath.Dir(path)
        if hasAncestorModule(root, moduleRoot, modules) { return nil }
        stack, ok := detectAt(moduleRoot)
        if !ok { return nil }
        rel, err := filepath.Rel(root, moduleRoot)
        if err != nil { return fmt.Errorf("resolve module path: %w", err) }
        modules = append(modules, Module{Path: filepath.ToSlash(rel), Stack: stack})
        return nil
    })
    if err != nil { return Workspace{}, err }
    sort.Slice(modules, func(i, j int) bool { return modules[i].Path < modules[j].Path })
    return Workspace{Root: root, Modules: modules}, nil
}

func detectAt(root string) (Stack, bool) {
    switch {
    case exists(root, "package.json"):
        stack := Stack{Name: "Node.js", Language: "javascript", PackageManager: detectNodeManager(root), DependencyFiles: nodeDependencyFiles(root), BuildCommand: "npm run build", StartCommand: "npm start", RuntimeCommand: "npm start"}
        if stack.PackageManager == "pnpm" { stack.BuildCommand, stack.StartCommand, stack.RuntimeCommand = "pnpm run build", "pnpm start", "pnpm start" }
        if stack.PackageManager == "yarn" { stack.BuildCommand, stack.StartCommand, stack.RuntimeCommand = "yarn build", "yarn start", "yarn start" }
        if build, start := nodeScripts(root); build != "" || start != "" {
            if build != "" { stack.BuildCommand = build }
            if start != "" { stack.StartCommand, stack.RuntimeCommand = start, start }
        }
        return stack, true
    case exists(root, "go.mod"):
        return Stack{Name: "Go", Language: "go", PackageManager: "go", DependencyFiles: []string{"go.mod", "go.sum"}, BuildCommand: "go build -trimpath -ldflags='-s -w' -o /out/autoline-app .", StartCommand: "/autoline-app", RuntimeCommand: "/autoline-app", BinaryName: "autoline-app"}, true
    case exists(root, "Cargo.toml"):
        name := cargoPackageName(root); if name == "" { name = "app" }
        return Stack{Name: "Rust", Language: "rust", PackageManager: "cargo", DependencyFiles: []string{"Cargo.toml", "Cargo.lock"}, BuildCommand: "cargo build --release", StartCommand: "/usr/local/bin/" + name, RuntimeCommand: "/usr/local/bin/" + name, BinaryName: name}, true
    case exists(root, "requirements.txt") || exists(root, "pyproject.toml"):
        deps := []string{}
        if exists(root, "requirements.txt") { deps = append(deps, "requirements.txt") }
        if exists(root, "pyproject.toml") { deps = append(deps, "pyproject.toml") }
        if exists(root, "poetry.lock") { deps = append(deps, "poetry.lock") }
        if exists(root, "uv.lock") { deps = append(deps, "uv.lock") }
        return Stack{Name: "Python", Language: "python", PackageManager: detectPythonManager(root), DependencyFiles: deps, BuildCommand: "python -m compileall .", StartCommand: "python app.py", RuntimeCommand: "python app.py"}, true
    case exists(root, "pom.xml"):
        return Stack{Name: "Java", Language: "java", PackageManager: "maven", DependencyFiles: []string{"pom.xml"}, BuildCommand: "./mvnw -B package -DskipTests || mvn -B package -DskipTests", StartCommand: "java -jar target/app.jar", RuntimeCommand: "java -jar target/app.jar"}, true
    case exists(root, "build.gradle") || exists(root, "build.gradle.kts") || exists(root, "settings.gradle") || exists(root, "settings.gradle.kts"):
        deps := []string{"build.gradle"}; if exists(root, "build.gradle.kts") { deps = []string{"build.gradle.kts"} }
        return Stack{Name: "Kotlin/Gradle", Language: "kotlin", PackageManager: "gradle", DependencyFiles: deps, BuildCommand: "./gradlew build || gradle build", StartCommand: "java -jar build/libs/app.jar", RuntimeCommand: "java -jar build/libs/app.jar"}, true
    case exists(root, "*.csproj") || exists(root, "*.sln"):
        return Stack{Name: "C#/.NET", Language: "csharp", PackageManager: "dotnet", DependencyFiles: []string{}, BuildCommand: "dotnet build --configuration Release", StartCommand: "dotnet run --configuration Release", RuntimeCommand: "dotnet run --configuration Release"}, true
    case exists(root, "composer.json"):
        return Stack{Name: "PHP", Language: "php", PackageManager: "composer", DependencyFiles: []string{"composer.json", "composer.lock"}, BuildCommand: "composer install --no-interaction --prefer-dist", StartCommand: "php -S 0.0.0.0:8080 -t public", RuntimeCommand: "php -S 0.0.0.0:8080 -t public"}, true
    case exists(root, "Gemfile"):
        return Stack{Name: "Ruby", Language: "ruby", PackageManager: "bundler", DependencyFiles: []string{"Gemfile", "Gemfile.lock"}, BuildCommand: "bundle install", StartCommand: "bundle exec ruby app.rb", RuntimeCommand: "bundle exec ruby app.rb"}, true
    case exists(root, "mix.exs"):
        return Stack{Name: "Elixir", Language: "elixir", PackageManager: "mix", DependencyFiles: []string{"mix.exs", "mix.lock"}, BuildCommand: "mix deps.get && mix compile", StartCommand: "mix run --no-halt", RuntimeCommand: "mix run --no-halt"}, true
    case exists(root, "pubspec.yaml"):
        return Stack{Name: "Dart/Flutter", Language: "dart", PackageManager: "pub", DependencyFiles: []string{"pubspec.yaml", "pubspec.lock"}, BuildCommand: "dart pub get && dart analyze", StartCommand: "dart run", RuntimeCommand: "dart run"}, true
    case exists(root, "Package.swift"):
        return Stack{Name: "Swift", Language: "swift", PackageManager: "swiftpm", DependencyFiles: []string{"Package.swift", "Package.resolved"}, BuildCommand: "swift build -c release", StartCommand: "swift run -c release", RuntimeCommand: "swift run -c release"}, true
    case exists(root, "CMakeLists.txt"):
        return Stack{Name: "C/C++", Language: "cpp", PackageManager: "cmake", DependencyFiles: []string{"CMakeLists.txt"}, BuildCommand: "cmake -S . -B build -DCMAKE_BUILD_TYPE=Release && cmake --build build --parallel", StartCommand: "./build/app", RuntimeCommand: "./build/app"}, true
    default:
        return Stack{}, false
    }
}

func isManifest(name string) bool {
    switch name {
    case "package.json", "go.mod", "Cargo.toml", "requirements.txt", "pyproject.toml", "pom.xml", "build.gradle", "build.gradle.kts", "settings.gradle", "settings.gradle.kts", "composer.json", "Gemfile", "mix.exs", "pubspec.yaml", "Package.swift", "CMakeLists.txt":
        return true
    }
    return strings.HasSuffix(name, ".csproj") || strings.HasSuffix(name, ".sln")
}

func isIgnoredDirectory(name string) bool {
    switch name {
    case ".git", ".hg", ".svn", "node_modules", "vendor", "target", ".venv", "venv", "dist", "build", ".cache", ".gradle", "bin", "obj":
        return true
    }
    return false
}

func hasAncestorModule(root, path string, modules []Module) bool {
    for _, module := range modules {
        modulePath := filepath.Join(root, filepath.FromSlash(module.Path))
        rel, err := filepath.Rel(modulePath, path)
        if err == nil && rel != "." && !strings.HasPrefix(rel, ".."+string(os.PathSeparator)) && rel != ".." { return true }
    }
    return false
}

func exists(root, name string) bool {
    if strings.ContainsAny(name, "*?[") {
        matches, _ := filepath.Glob(filepath.Join(root, name))
        return len(matches) > 0
    }
    _, err := os.Stat(filepath.Join(root, name))
    return err == nil
}

func detectNodeManager(root string) string {
    switch {
    case exists(root, "pnpm-lock.yaml"): return "pnpm"
    case exists(root, "yarn.lock"): return "yarn"
    case exists(root, "bun.lockb") || exists(root, "bun.lock"): return "bun"
    default: return "npm"
    }
}

func nodeDependencyFiles(root string) []string {
    files := []string{"package.json"}
    for _, name := range []string{"package-lock.json", "pnpm-lock.yaml", "yarn.lock", "bun.lockb", "bun.lock"} {
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
    data, err := os.ReadFile(filepath.Join(root, "package.json")); if err != nil { return "", "" }
    var manifest struct { Scripts map[string]string `json:"scripts"` }
    if err := json.Unmarshal(data, &manifest); err != nil { return "", "" }
    build, start := "", ""
    if _, ok := manifest.Scripts["build"]; ok { build = "npm run build" }
    if _, ok := manifest.Scripts["start"]; ok { start = "npm start" }
    return build, start
}

func cargoPackageName(root string) string {
    data, err := os.ReadFile(filepath.Join(root, "Cargo.toml")); if err != nil { return "" }
    inPackage := false
    for _, line := range strings.Split(string(data), "\n") {
        line = strings.TrimSpace(line)
        if line == "[package]" { inPackage = true; continue }
        if len(line) > 0 && line[0] == '[' { inPackage = false }
        if inPackage && strings.HasPrefix(line, "name") {
            parts := strings.SplitN(line, "=", 2)
            if len(parts) == 2 { return strings.Trim(strings.TrimSpace(parts[1]), "\"'") }
        }
    }
    return ""
}
