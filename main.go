package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func run(args []string, cwd string, input io.Reader, stdout, stderr io.Writer) int {
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
		filename, err = findConfig(cwd)
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 2
		}
	} else if !filepath.IsAbs(filename) {
		filename = filepath.Join(cwd, filename)
	}

	cfg, err := readConfig(filename)
	if err != nil {
		fmt.Fprintf(stderr, "%s: %v\n", filename, err)
		return 2
	}
	edges, err := readGraph(input)
	if err != nil {
		fmt.Fprintf(stderr, "graph: %v\n", err)
		return 2
	}
	if *referenceFile != "" {
		return checkReferences(cfg, edges, *referenceFile, stdout, stderr)
	}

	layers := make(map[string]int)
	layer := func(file string) int {
		if index, known := layers[file]; known {
			return index
		}
		index := cfg.layer(file)
		layers[file] = index
		if cfg.debug {
			fmt.Fprintf(stderr, "processing file %s\n", file)
		}
		return index
	}
	violations := []violation{}
	for _, e := range edges {
		from, to := layer(e.From), layer(e.To)
		if from >= 0 && from < to {
			violations = append(violations, violation{e, cfg.layers[from], cfg.layers[to]})
		}
	}
	if err := json.NewEncoder(stdout).Encode(violations); err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}
	if len(violations) > 0 {
		return 1
	}
	return 0
}

func main() {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	os.Exit(run(os.Args[1:], cwd, os.Stdin, os.Stdout, os.Stderr))
}
