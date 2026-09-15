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
    if err != nil {
        return Stack{}, err
    }
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
    if err != nil {
        return Workspace{}, fmt.Errorf("inspect repository: %w", err)
    }
    if !info.IsDir() {
        return Workspace{}, fmt.Errorf("repository path is not a directory: %s", root)
    }

    modules := make([]Module, 0, 8)
    // Create a map for O(1) ancestor lookups by path prefix
    modulePathSet := make(map[string]bool)

    err = filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
        if walkErr != nil {
            return fmt.Errorf("walk %s: %w", path, walkErr)
        }
        if entry.IsDir() {
            if path != root && isIgnoredDirectory(entry.Name()) {
                return filepath.SkipDir
            }
            return nil
        }
        if !isManifest(filepath.Base(path)) {
            return nil
        }
        moduleRoot := filepath.Dir(path)
        // Use path-based lookup instead of linear iteration
        if hasAncestorModuleOptimized(root, moduleRoot, modulePathSet) {
            return nil
        }
        stack, ok := detectAt(moduleRoot)
        if !ok {
            return nil
        }
        rel, err := filepath.Rel(root, moduleRoot)
        if err != nil {
            return fmt.Errorf("resolve module path: %w", err)
        }
        relSlash := filepath.ToSlash(rel)
        modules = append(modules, Module{Path: relSlash, Stack: stack})
        modulePathSet[moduleRoot] = true
        return nil
    })
    if err != nil {
        return Workspace{}, err
    }
    sort.Slice(modules, func(i, j int) bool {
        return modules[i].Path < modules[j].Path
    })
    return Workspace{Root: root, Modules: modules}, nil
}

// hasAncestorModuleOptimized checks if moduleRoot is nested within an already-detected module.
// Uses a map lookup instead of iterating through all modules (O(1) vs O(n)).
func hasAncestorModuleOptimized(root, moduleRoot string, modulePathSet map[string]bool) bool {
    current := moduleRoot
    for current != root && current != filepath.Dir(current) {
        if modulePathSet[current] {
            return true
        }
        current = filepath.Dir(current)
    }
    return false
}

func detectAt(root string) (Stack, error) {
    // Batch all file checks upfront by reading the directory once
    entries, err := os.ReadDir(root)
    if err != nil {
        return Stack{}, fmt.Errorf("read directory %s: %w", root, err)
    }
    fileSet := make(map[string]bool)
    for _, entry := range entries {
        if !entry.IsDir() {
            fileSet[entry.Name()] = true
        }
    }

    switch {
    case fileSet["package.json"]:
        stack := Stack{Name: "Node.js", Language: "javascript", PackageManager: detectNodeManagerFromSet(fileSet), DependencyFiles: nodeDependencyFilesFromSet(fileSet), BuildCommand: "npm run build", StartCommand: "npm start", RuntimeCommand: "npm start"}
        if stack.PackageManager == "pnpm" {
            stack.BuildCommand, stack.StartCommand, stack.RuntimeCommand = "pnpm run build", "pnpm start", "pnpm start"
        }
        if stack.PackageManager == "yarn" {
            stack.BuildCommand, stack.StartCommand, stack.RuntimeCommand = "yarn build", "yarn start", "yarn start"
        }
        if build, start := nodeScripts(root); build != "" || start != "" {
            if build != "" {
                stack.BuildCommand = build
            }
            if start != "" {
                stack.StartCommand, stack.RuntimeCommand = start, start
            }
        }
        return stack, nil
    case fileSet["go.mod"]:
        return Stack{Name: "Go", Language: "go", PackageManager: "go", DependencyFiles: []string{"go.mod", "go.sum"}, BuildCommand: "go build -trimpath -ldflags='-s -w' -o /out/autoline-app .", StartCommand: "/out/autoline-app", RuntimeCommand: "/out/autoline-app"}, nil
    case fileSet["Cargo.toml"]:
        name := cargoPackageNameOptimized(root)
        if name == "" {
            name = "app"
        }
        return Stack{Name: "Rust", Language: "rust", PackageManager: "cargo", DependencyFiles: []string{"Cargo.toml", "Cargo.lock"}, BuildCommand: "cargo build --release", StartCommand: "/usr/local/cargo/bin/" + name, RuntimeCommand: "/usr/local/cargo/bin/" + name, BinaryName: name}, nil
    case fileSet["requirements.txt"] || fileSet["pyproject.toml"]:
        deps := []string{}
        if fileSet["requirements.txt"] {
            deps = append(deps, "requirements.txt")
        }
        if fileSet["pyproject.toml"] {
            deps = append(deps, "pyproject.toml")
        }
        if fileSet["poetry.lock"] {
            deps = append(deps, "poetry.lock")
        }
        if fileSet["uv.lock"] {
            deps = append(deps, "uv.lock")
        }
        return Stack{Name: "Python", Language: "python", PackageManager: detectPythonManagerFromSet(fileSet), DependencyFiles: deps, BuildCommand: "python -m compileall .", StartCommand: "python app.py", RuntimeCommand: "python app.py"}, nil
    case fileSet["pom.xml"]:
        return Stack{Name: "Java", Language: "java", PackageManager: "maven", DependencyFiles: []string{"pom.xml"}, BuildCommand: "./mvnw -B package -DskipTests || mvn -B package -DskipTests", StartCommand: "java -jar target/app.jar", RuntimeCommand: "java -jar target/app.jar"}, nil
    case fileSet["build.gradle"] || fileSet["build.gradle.kts"] || fileSet["settings.gradle"] || fileSet["settings.gradle.kts"]:
        deps := []string{"build.gradle"}
        if fileSet["build.gradle.kts"] {
            deps = []string{"build.gradle.kts"}
        }
        return Stack{Name: "Kotlin/Gradle", Language: "kotlin", PackageManager: "gradle", DependencyFiles: deps, BuildCommand: "./gradlew build || gradle build", StartCommand: "java -jar build/libs/*.jar", RuntimeCommand: "java -jar build/libs/*.jar"}, nil
    case hasCsprojOrSln(fileSet):
        return Stack{Name: "C#/.NET", Language: "csharp", PackageManager: "dotnet", DependencyFiles: []string{}, BuildCommand: "dotnet build --configuration Release", StartCommand: "dotnet run --no-build", RuntimeCommand: "dotnet run --no-build"}, nil
    case fileSet["composer.json"]:
        return Stack{Name: "PHP", Language: "php", PackageManager: "composer", DependencyFiles: []string{"composer.json", "composer.lock"}, BuildCommand: "composer install --no-interaction --prefer-dist", StartCommand: "php -S localhost:8000", RuntimeCommand: "php -S localhost:8000"}, nil
    case fileSet["Gemfile"]:
        return Stack{Name: "Ruby", Language: "ruby", PackageManager: "bundler", DependencyFiles: []string{"Gemfile", "Gemfile.lock"}, BuildCommand: "bundle install", StartCommand: "bundle exec ruby app.rb", RuntimeCommand: "bundle exec ruby app.rb"}, nil
    case fileSet["mix.exs"]:
        return Stack{Name: "Elixir", Language: "elixir", PackageManager: "mix", DependencyFiles: []string{"mix.exs", "mix.lock"}, BuildCommand: "mix deps.get && mix compile", StartCommand: "mix run", RuntimeCommand: "mix run"}, nil
    case fileSet["pubspec.yaml"]:
        return Stack{Name: "Dart/Flutter", Language: "dart", PackageManager: "pub", DependencyFiles: []string{"pubspec.yaml", "pubspec.lock"}, BuildCommand: "dart pub get && dart analyze", StartCommand: "dart run", RuntimeCommand: "dart run"}, nil
    case fileSet["Package.swift"]:
        return Stack{Name: "Swift", Language: "swift", PackageManager: "swiftpm", DependencyFiles: []string{"Package.swift", "Package.resolved"}, BuildCommand: "swift build -c release", StartCommand: "swift run", RuntimeCommand: "swift run"}, nil
    case fileSet["CMakeLists.txt"]:
        return Stack{Name: "C/C++", Language: "cpp", PackageManager: "cmake", DependencyFiles: []string{"CMakeLists.txt"}, BuildCommand: "cmake -S . -B build -DCMAKE_BUILD_TYPE=Release && cmake --build build", StartCommand: "./build/app", RuntimeCommand: "./build/app"}, nil
    default:
        return Stack{}, false
    }
}

func hasCsprojOrSln(fileSet map[string]bool) bool {
    for name := range fileSet {
        if strings.HasSuffix(name, ".csproj") || strings.HasSuffix(name, ".sln") {
            return true
        }
    }
    return false
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

func detectNodeManagerFromSet(fileSet map[string]bool) string {
    switch {
    case fileSet["pnpm-lock.yaml"]:
        return "pnpm"
    case fileSet["yarn.lock"]:
        return "yarn"
    case fileSet["bun.lockb"] || fileSet["bun.lock"]:
        return "bun"
    default:
        return "npm"
    }
}

func nodeDependencyFilesFromSet(fileSet map[string]bool) []string {
    files := []string{"package.json"}
    for _, name := range []string{"package-lock.json", "pnpm-lock.yaml", "yarn.lock", "bun.lockb", "bun.lock"} {
        if fileSet[name] {
            files = append(files, name)
        }
    }
    return files
}

func detectPythonManagerFromSet(fileSet map[string]bool) string {
    if fileSet["uv.lock"] {
        return "uv"
    }
    if fileSet["poetry.lock"] {
        return "poetry"
    }
    return "pip"
}

func nodeScripts(root string) (string, error) {
    data, err := os.ReadFile(filepath.Join(root, "package.json"))
    if err != nil {
        return "", ""
    }
    var manifest struct {
        Scripts map[string]string `json:"scripts"`
    }
    if err := json.Unmarshal(data, &manifest); err != nil {
        return "", ""
    }
    build, start := "", ""
    if _, ok := manifest.Scripts["build"]; ok {
        build = "npm run build"
    }
    if _, ok := manifest.Scripts["start"]; ok {
        start = "npm start"
    }
    return build, start
}

// cargoPackageNameOptimized reads Cargo.toml once and exits early after finding package name.
func cargoPackageNameOptimized(root string) string {
    data, err := os.ReadFile(filepath.Join(root, "Cargo.toml"))
    if err != nil {
        return ""
    }
    inPackage := false
    for _, line := range strings.Split(string(data), "\n") {
        line = strings.TrimSpace(line)
        if line == "[package]" {
            inPackage = true
            continue
        }
        // Early exit once we leave the [package] block
        if len(line) > 0 && line[0] == '[' {
            if inPackage {
                return "" // Hit next section, package name wasn't found
            }
        }
        if inPackage && strings.HasPrefix(line, "name") {
            parts := strings.SplitN(line, "=", 2)
            if len(parts) == 2 {
                return strings.Trim(strings.TrimSpace(parts[1]), "\"'")
            }
        }
    }
    return ""
}
