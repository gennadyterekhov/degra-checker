package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"slices"
)

// A reference can describe a resolved edge or an import whose target has no
// file in the selected graph, such as a Go package outside the selection.
type reference struct {
	edge
	ResolvedLayers bool   `json:"resolvedLayers"`
	FromLayer      string `json:"fromLayer"`
	ToLayer        string `json:"toLayer"`
}

type referenceViolation struct {
	Index     int    `json:"index"`
	FromLayer string `json:"fromLayer"`
	ToLayer   string `json:"toLayer"`
}

func checkReferences(cfg config, edges []edge, filename string, stdout, stderr io.Writer) int {
	data, err := os.ReadFile(filename)
	var refs []reference
	if err == nil {
		err = json.Unmarshal(data, &refs)
	}
	if err != nil {
		fmt.Fprintf(stderr, "references: %v\n", err)
		return 2
	}
	graph := make(map[edge]bool, len(edges))
	for _, e := range edges {
		graph[e] = true
	}
	violations := []referenceViolation{}
	for i, ref := range refs {
		if ref.To != "" && !graph[ref.edge] {
			continue
		}
		from, to := slices.Index(cfg.layers, ref.FromLayer), slices.Index(cfg.layers, ref.ToLayer)
		if !ref.ResolvedLayers {
			from, to = cfg.layer(ref.From), cfg.layer(ref.To)
		}
		if from >= 0 && from < to {
			violations = append(violations, referenceViolation{i, cfg.layers[from], cfg.layers[to]})
		}
	}
	if err := json.NewEncoder(stdout).Encode(violations); err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}
	if len(violations) != 0 {
		return 1
	}
	return 0
}
