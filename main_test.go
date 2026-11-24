package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCheck(t *testing.T) {
	grouped := `{
		"src/domain/main.go":["src/app/run.go","src/app/run.go","src/domain/other.go","src/domain-extra/file.go","other/app/file.go","src/app/run_test.go"],
		"src/app/run.go":["src/domain/main.go"],
		"src/domain-extra/file.go":["src/app/run.go"],
		"src/domain/main_test.go":["src/app/run.go"]
	}`
	edges := `[{"from":"src/domain/main.go","to":"src/app/run.go"},{"from":"src/app/run.go","to":"src/domain/main.go"}]`
	for _, tt := range []struct{ name, config, graph string }{
		{"yaml", "layers: [src/domain, app]\nrootDir: src\nignoreTests: true\n", grouped},
		{"json", `{"layers":["src/domain","app"],"rootDir":"src","ignore_tests":true}`, grouped},
		{"edges", `{"layers":["src/domain","app"]}`, edges},
	} {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			ext := ".json"
			if tt.name == "yaml" {
				ext = ".yaml"
			}
			config := filepath.Join(dir, "config"+ext)
			if err := os.WriteFile(config, []byte(tt.config), 0600); err != nil {
				t.Fatal(err)
			}
			var stdout, stderr bytes.Buffer
			code := run([]string{"-c", config}, dir, strings.NewReader(tt.graph), &stdout, &stderr)
			want := "[{\"from\":\"src/domain/main.go\",\"to\":\"src/app/run.go\",\"fromLayer\":\"src/domain\",\"toLayer\":\"app\"}]\n"
			if code != 1 || stdout.String() != want || stderr.Len() != 0 {
				t.Fatalf("code %d, stdout %q, stderr %q", code, stdout.String(), stderr.String())
			}
		})
	}
}

func TestCheckPathPriorityAndPythonModules(t *testing.T) {
	dir := t.TempDir()
	config := `{"layers":["domain","src/domain/special","analyzer.layer_policy","analyzer"]}`
	if err := os.WriteFile(filepath.Join(dir, "import_layers.json"), []byte(config), 0600); err != nil {
		t.Fatal(err)
	}
	graph := `{"src/domain/b.py":["src/domain/special/a.py"],"src/pkg/analyzer/layer_policy.py":["src/pkg/analyzer/main.py"]}`
	var stdout, stderr bytes.Buffer
	code := run(nil, dir, strings.NewReader(graph), &stdout, &stderr)
	want := `[{"from":"src/domain/b.py","to":"src/domain/special/a.py","fromLayer":"domain","toLayer":"src/domain/special"},{"from":"src/pkg/analyzer/layer_policy.py","to":"src/pkg/analyzer/main.py","fromLayer":"analyzer.layer_policy","toLayer":"analyzer"}]` + "\n"
	if code != 1 || stdout.String() != want || stderr.Len() != 0 {
		t.Fatalf("code %d, stdout %q, stderr %q", code, stdout.String(), stderr.String())
	}
}

func TestCleanAndInvalid(t *testing.T) {
	dir := t.TempDir()
	for _, tt := range []struct {
		config, graph string
		code          int
	}{
		{`{"layers":["domain","app"]}`, `{"app/a.js":["domain/b.js"]}`, 0},
		{`{"layers":["domain","domain"]}`, `{}`, 2},
		{`{"layers":["domain"]}`, `{"domain/a.go":null}`, 2},
		{`{"layers":["domain"]}`, `{} {}`, 2},
	} {
		if err := os.WriteFile(filepath.Join(dir, "import_layers.json"), []byte(tt.config), 0600); err != nil {
			t.Fatal(err)
		}
		var stdout, stderr bytes.Buffer
		code := run(nil, dir, strings.NewReader(tt.graph), &stdout, &stderr)
		if code != tt.code || code == 0 && (stdout.String() != "[]\n" || stderr.Len() != 0) ||
			code == 2 && (stdout.Len() != 0 || stderr.Len() == 0) {
			t.Errorf("config %s, graph %s: code %d, stdout %q, stderr %q", tt.config, tt.graph, code, stdout.String(), stderr.String())
		}
	}
}
