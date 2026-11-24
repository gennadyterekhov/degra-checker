package main

import (
	"path"
	"slices"
	"strings"
)

func (cfg config) layer(file string) int {
	dir := path.Dir(file)
	if !within(dir, cfg.rootDir) || cfg.ignoreTests && isTestFile(file) {
		return -1
	}
	for i, name := range cfg.layers {
		if strings.Contains(name, "/") && within(dir, name) {
			return i
		}
	}

	segments := strings.Split(dir, "/")
	if strings.HasSuffix(file, ".py") {
		module := strings.TrimPrefix(file, "src/")
		module = strings.TrimSuffix(module, ".py")
		segments = strings.Split(module, "/")
		if segments[len(segments)-1] == "__init__" {
			segments = segments[:len(segments)-1]
		}
	}
	for i, name := range cfg.layers {
		if strings.Contains(name, "/") {
			continue
		}
		if strings.HasSuffix(file, ".py") {
			parts := strings.Split(name, ".")
			for start := 0; start+len(parts) <= len(segments); start++ {
				if slices.Equal(segments[start:start+len(parts)], parts) {
					return i
				}
			}
		} else if slices.Contains(segments, name) {
			return i
		}
	}
	return -1
}

func isTestFile(file string) bool {
	ext := path.Ext(file)
	name := strings.TrimSuffix(path.Base(file), ext)
	if ext == ".go" {
		return strings.HasSuffix(name, "_test")
	}
	for _, dir := range strings.Split(path.Dir(file), "/") {
		if dir == "test" || dir == "tests" || dir == "__tests__" {
			return true
		}
	}
	switch ext {
	case ".py":
		return strings.HasPrefix(name, "test_") || strings.HasSuffix(name, "_test")
	case ".php":
		return strings.HasSuffix(name, "Test") || strings.HasSuffix(name, "_test")
	default:
		return strings.HasSuffix(name, ".test") || strings.HasSuffix(name, ".spec")
	}
}
