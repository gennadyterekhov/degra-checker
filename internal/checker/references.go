package checker

import "slices"

// A Reference can describe a resolved edge or an import whose target has no
// file in the selected graph, such as a Go package outside the selection.
type Reference struct {
	Edge
	ResolvedLayers bool   `json:"resolvedLayers"`
	FromLayer      string `json:"fromLayer"`
	ToLayer        string `json:"toLayer"`
}

type referenceViolation struct {
	Index     int    `json:"index"`
	FromLayer string `json:"fromLayer"`
	ToLayer   string `json:"toLayer"`
}

func (cfg Config) CheckReferences(edges []Edge, refs []Reference) []referenceViolation {
	graph := make(map[Edge]bool, len(edges))
	for _, e := range edges {
		graph[e] = true
	}
	violations := []referenceViolation{}
	for i, ref := range refs {
		if ref.To != "" && !graph[ref.Edge] {
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
	return violations
}
