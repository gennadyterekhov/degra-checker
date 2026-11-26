package checker

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	layers      []string
	rootDir     string
	ignoreTests bool
	debug       bool
}

func FindConfig(cwd string) (string, error) {
	for dir := cwd; ; dir = filepath.Dir(dir) {
		for _, name := range []string{"import_layers.json", "import_layers.yaml", "import_layers.yml"} {
			filename := filepath.Join(dir, name)
			info, err := os.Stat(filename)
			if err == nil && info.Mode().IsRegular() {
				return filename, nil
			}
			if err != nil && !errors.Is(err, os.ErrNotExist) {
				return "", err
			}
		}
		if filepath.Dir(dir) == dir {
			return "", errors.New("import_layers.json, import_layers.yaml or import_layers.yml not found")
		}
	}
}

func ReadConfig(filename string) (Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return Config{}, err
	}
	var fields map[string]any
	switch strings.ToLower(filepath.Ext(filename)) {
	case ".yaml", ".yml":
		err = yaml.Unmarshal(data, &fields)
	default:
		err = json.Unmarshal(data, &fields)
	}
	if err != nil {
		return Config{}, err
	}

	var cfg Config
	names, ok := fields["layers"].([]any)
	if !ok || len(names) == 0 {
		return Config{}, errors.New("layers must be a non-empty list of unique strings")
	}
	seen := make(map[string]bool)
	for _, value := range names {
		name, ok := value.(string)
		if !ok || strings.TrimSpace(name) == "" || seen[name] {
			return Config{}, errors.New("layers must be a non-empty list of unique strings")
		}
		seen[name] = true
		cfg.layers = append(cfg.layers, name)
	}

	if value, present := fields["rootDir"]; present {
		dir, ok := value.(string)
		if !ok || path.IsAbs(dir) || !within(path.Clean(dir), ".") || strings.ContainsRune(dir, 0) {
			return Config{}, errors.New("rootDir must be a project-relative directory path")
		}
		cfg.rootDir = path.Clean(dir)
	}
	ignoreTests := "ignoreTests"
	if _, present := fields[ignoreTests]; !present {
		ignoreTests = "ignore_tests"
	}
	cfg.ignoreTests, err = configBool(fields, ignoreTests)
	if err != nil {
		return Config{}, err
	}
	cfg.debug, err = configBool(fields, "debug")
	return cfg, err
}

func configBool(fields map[string]any, key string) (bool, error) {
	value, present := fields[key]
	if !present {
		return false, nil
	}
	b, ok := value.(bool)
	if !ok {
		return false, fmt.Errorf("%s must be a boolean", key)
	}
	return b, nil
}
