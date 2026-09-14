package detector

import (
    "fmt"
    "os"
    "path/filepath"
)

type Stack struct {
    Name string
    Language string
    PackageManager string
    BuildCommand string
    StartCommand string
}

func Detect(root string) (Stack, error) {
    info, err := os.Stat(root)
    if err != nil { return Stack{}, fmt.Errorf("inspect repository: %w", err) }
    if !info.IsDir() { return Stack{}, fmt.Errorf("repository path is not a directory: %s", root) }
    checks := []struct{ file string; stack Stack }{
        {"package.json", Stack{"Node.js", "javascript", "npm", "npm run build", "npm start"}},
        {"requirements.txt", Stack{"Python", "python", "pip", "python -m compileall .", "python app.py"}},
        {"pyproject.toml", Stack{"Python", "python", "pip", "python -m compileall .", "python app.py"}},
        {"go.mod", Stack{"Go", "go", "go", "go build -o app .", "./app"}},
        {"Cargo.toml", Stack{"Rust", "rust", "cargo", "cargo build --release", "./target/release/app"}},
    }
    for _, c := range checks {
        if _, err := os.Stat(filepath.Join(root, c.file)); err == nil { return c.stack, nil }
        if !os.IsNotExist(err) { return Stack{}, fmt.Errorf("check %s: %w", c.file, err) }
    }
    return Stack{Name: "Generic", Language: "unknown", PackageManager: "unknown", BuildCommand: "echo 'No build detected'", StartCommand: "echo 'No start command detected'"}, nil
}
