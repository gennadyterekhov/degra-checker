package cli

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/gennadyterekhov/degra-checker/internal/checker"
)

func Run(args []string, cwd string, input io.Reader, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("degra-checker", flag.ContinueOnError)
	flags.SetOutput(stderr)
	flags.Usage = func() {
		fmt.Fprintln(stderr, "usage: degra-checker [-c|--config PATH] < graph.json")
	}
	var filename string
	referenceFile := flags.String("references", "", "check language-resolved import references")
	flags.StringVar(&filename, "c", "", "layer configuration")
	flags.StringVar(&filename, "config", "", "layer configuration")
	if err := flags.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return 0
		}
		return 2
	}
	if flags.NArg() != 0 {
		flags.Usage()
		return 2
	}
	if filename == "" {
		var err error
		filename, err = checker.FindConfig(cwd)
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 2
		}
	} else if !filepath.IsAbs(filename) {
		filename = filepath.Join(cwd, filename)
	}

	cfg, err := checker.ReadConfig(filename)
	if err != nil {
		fmt.Fprintf(stderr, "%s: %v\n", filename, err)
		return 2
	}
	edges, err := checker.ReadGraph(input)
	if err != nil {
		fmt.Fprintf(stderr, "graph: %v\n", err)
		return 2
	}
	if *referenceFile != "" {
		data, err := os.ReadFile(*referenceFile)
		var refs []checker.Reference
		if err == nil {
			err = json.Unmarshal(data, &refs)
		}
		if err != nil {
			fmt.Fprintf(stderr, "references: %v\n", err)
			return 2
		}
		return outputViolations(cfg.CheckReferences(edges, refs), stdout, stderr)
	}

	violations := cfg.Check(edges, func(file string) {
		fmt.Fprintf(stderr, "processing file %s\n", file)
	})
	return outputViolations(violations, stdout, stderr)
}

func outputViolations[T any](violations []T, stdout, stderr io.Writer) int {
	if err := json.NewEncoder(stdout).Encode(violations); err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}
	if len(violations) > 0 {
		return 1
	}
	return 0
}
