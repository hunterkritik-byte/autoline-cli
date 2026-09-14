package detector

import (
    "os"
    "path/filepath"
    "testing"
)

func TestDetect(t *testing.T) {
    tests := []struct {
        name, marker, wantName, wantManager string
    }{
        {"node npm", "package.json", "Node.js", "npm"},
        {"node pnpm", "pnpm-lock.yaml", "Node.js", "pnpm"},
        {"python", "requirements.txt", "Python", "pip"},
        {"go", "go.mod", "Go", "go"},
        {"rust", "Cargo.toml", "Rust", "cargo"},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            root := t.TempDir()
            if tt.wantName == "Node.js" && tt.marker != "package.json" {
                if err := os.WriteFile(filepath.Join(root, "package.json"), []byte(`{"name":"demo"}`), 0o644); err != nil { t.Fatal(err) }
            }
            content := []byte("name = \"demo\"\n")
            if tt.marker == "package.json" { content = []byte(`{"name":"demo"}`) }
            if err := os.WriteFile(filepath.Join(root, tt.marker), content, 0o644); err != nil { t.Fatal(err) }
            stack, err := Detect(root)
            if err != nil { t.Fatal(err) }
            if stack.Name != tt.wantName || stack.PackageManager != tt.wantManager { t.Fatalf("got %s/%s, want %s/%s", stack.Name, stack.PackageManager, tt.wantName, tt.wantManager) }
        })
    }
}

func TestDetectNodeScripts(t *testing.T) {
    root := t.TempDir()
    manifest := `{"scripts":{"build":"vite build","start":"node server.js"}}`
    if err := os.WriteFile(filepath.Join(root, "package.json"), []byte(manifest), 0o644); err != nil { t.Fatal(err) }
    stack, err := Detect(root)
    if err != nil { t.Fatal(err) }
    if stack.BuildCommand != "npm run build" || stack.StartCommand != "npm start" { t.Fatalf("unexpected commands: %q %q", stack.BuildCommand, stack.StartCommand) }
}
